package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	vmedisv2 "github.com/turfaa/vmedis-proxy-api/vmedis/v2"
)

var journalsCmd = &cobra.Command{
	Use:   "journals",
	Short: "Journal entry (jurnal keuangan) commands",
}

var journalsCommands = []commandWithInit{
	{
		command: &cobra.Command{
			Use:   "smoke-test",
			Short: "Create a test journal entry of Rp1 in vmedis, look it up, and delete it",
			Long: `Creates a test journal entry of Rp1 (Bank OCBC to Gaji Karyawan by default)
dated --date, with a "TEST HAPUS" keterangan, under the configured vmedis user. It
then checks that the No. Bukti uses that date, that the entry is listed with and
without the keterangan filter of vmedis, and that a reused No. Bukti is
rejected, and finally deletes the entry (unless --keep).`,
			Run: func(cmd *cobra.Command, args []string) {
				dateStr, _ := cmd.Flags().GetString("date")
				credit, _ := cmd.Flags().GetString("credit")
				debit, _ := cmd.Flags().GetString("debit")
				keep, _ := cmd.Flags().GetBool("keep")

				if err := runJournalSmokeTest(cmd, dateStr, credit, debit, keep); err != nil {
					log.Fatalf("Journal smoke test failed: %s", err)
				}
			},
		},
		init: func(cmd *cobra.Command) {
			lastDayOfLastMonth := time.Now().AddDate(0, 0, -time.Now().Day())
			cmd.Flags().String("date", lastDayOfLastMonth.Format(time.DateOnly), "date of the test entry")
			cmd.Flags().String("credit", "111004", "credit account code")
			cmd.Flags().String("debit", "610002", "debit account code")
			cmd.Flags().Bool("keep", false, "keep the test entry instead of deleting it")
		},
	},
}

func runJournalSmokeTest(cmd *cobra.Command, dateStr, creditCode, debitCode string, keep bool) error {
	ctx := cmd.Context()
	client := getVmedisV2Client()

	date, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		return fmt.Errorf("parse date: %w", err)
	}

	accounts, err := client.ListJournalAccounts(ctx)
	if err != nil {
		return err
	}
	ids := map[string]int{}
	for _, a := range accounts {
		ids[a.Code] = a.ID
	}
	if ids[creditCode] == 0 || ids[debitCode] == 0 {
		return fmt.Errorf("account %s or %s not found among %d accounts", creditCode, debitCode, len(accounts))
	}
	log.Printf("OK accounts: %s = %d, %s = %d", creditCode, ids[creditCode], debitCode, ids[debitCode])

	voucherNo, err := client.ReserveJournalVoucherNo(ctx, date)
	if err != nil {
		return err
	}
	wantPrefix := "JRLM" + date.Format("060102")
	if strings.HasPrefix(voucherNo, wantPrefix) {
		log.Printf("OK No. Bukti %s uses the date", voucherNo)
	} else {
		log.Printf("CHECK No. Bukti %s does not start with %s", voucherNo, wantPrefix)
	}

	description := fmt.Sprintf("TEST HAPUS - smoke test jurnal %s", time.Now().Format("20060102150405"))
	created, err := client.CreateJournalEntry(ctx, vmedisv2.NewJournalEntry{
		Date:            date,
		VoucherNo:       voucherNo,
		Description:     description,
		DebitAccountID:  ids[debitCode],
		CreditAccountID: ids[creditCode],
		Amount:          1,
	})
	if err != nil {
		return fmt.Errorf("create the test entry (check vmedis for %q): %w", description, err)
	}
	log.Printf("OK created %s (jrlid %d), %q", created.VoucherNo, created.ID, description)

	defer func() {
		if keep {
			log.Printf("Kept %s (jrlid %d); delete it in vmedis", created.VoucherNo, created.ID)
			return
		}
		if err := client.DeleteJournalEntry(ctx, created.ID); err != nil {
			log.Printf("FAIL deleting %s (jrlid %d), delete it in vmedis: %s", created.VoucherNo, created.ID, err)
			return
		}
		log.Printf("OK deleted %s", created.VoucherNo)
	}()

	for _, filter := range []string{"", description} {
		entries, err := client.FindJournalEntries(ctx, vmedisv2.JournalEntryFilter{From: date, Until: date, Description: filter})
		if err != nil {
			return err
		}

		var found *vmedisv2.JournalEntry
		for i, e := range entries {
			if e.ID == created.ID {
				found = &entries[i]
			}
		}

		switch {
		case found == nil:
			log.Printf("FAIL listing with keterangan filter %q: %d entries, the test entry is not among them", filter, len(entries))
		default:
			log.Printf("OK listing with keterangan filter %q: %d entries, the test entry is listed under user %q at %s", filter, len(entries), found.Username, found.Time)
		}
	}

	reused, err := client.CreateJournalEntry(ctx, vmedisv2.NewJournalEntry{
		Date:            date,
		VoucherNo:       voucherNo,
		Description:     description + " (No. Bukti ganda)",
		DebitAccountID:  ids[debitCode],
		CreditAccountID: ids[creditCode],
		Amount:          1,
	})
	if err != nil {
		log.Printf("OK a reused No. Bukti is rejected: %s", err)
	} else {
		log.Printf("CHECK a reused No. Bukti is accepted, as %s (jrlid %d)", reused.VoucherNo, reused.ID)
		if !keep {
			if err := client.DeleteJournalEntry(ctx, reused.ID); err != nil {
				log.Printf("FAIL deleting %s (jrlid %d), delete it in vmedis: %s", reused.VoucherNo, reused.ID, err)
			} else {
				log.Printf("OK deleted %s (jrlid %d)", reused.VoucherNo, reused.ID)
			}
		}
	}

	return nil
}

func init() {
	initSubcommands(journalsCmd, journalsCommands)
}
