package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// DrugsResponse is the response of the Drugs client method.
type DrugsResponse struct {
	Drugs      []Drug
	OtherPages []int
}

// GetAllDrugs gets all the drugs from vmedis.
// It starts with getting the number of pages by calling the API with page 9999. The last page is the number of pages.
// Then it calls the /obat-batch/index?page=<page> page and try to parse the drugs from it.
// The resulting drugs will be returned in a channel.
func (c *Client) GetAllDrugs(ctx context.Context) (<-chan Drug, error) {
	// Get the number of pages
	log.Println("Getting number of pages of drugs")
	res, err := c.GetDrugs(ctx, 9999)
	if err != nil {
		return nil, fmt.Errorf("get number of pages: %w", err)
	}

	lastPage := 0
	for _, p := range res.OtherPages {
		if p > lastPage {
			lastPage = p
		}
	}

	log.Printf("Number of drugs pages: %d\n", lastPage)

	var (
		drugs = make(chan Drug, lastPage*10)
		pages = make(chan int, c.concurrency*2)
		wg    sync.WaitGroup
	)

	go func() {
		for i := 1; i <= lastPage; i++ {
			pages <- i
		}
		close(pages)
	}()

	// Start the workers
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for page := range pages {
				log.Printf("Getting drugs from page %d\n", page)

				res, err := c.GetDrugs(ctx, page)
				if err != nil {
					log.Printf("error getting drugs from page %d: %s", page, err)
					continue
				}

				log.Printf("Got %d drugs from page %d\n", len(res.Drugs), page)
				for _, drug := range res.Drugs {
					drugs <- drug
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(drugs)
	}()

	return drugs, nil
}

// GetDrugs gets the drugs from one page of "Data Obat" page in vmedis.
// It calls the /obat-batch/index?page=<page> page and try to parse the drugs from it.
func (c *Client) GetDrugs(ctx context.Context, page int) (DrugsResponse, error) {
	res, err := c.get(ctx, fmt.Sprintf("/obat-batch/index?page=%d", page))
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("get drugs: %w", err)
	}
	defer res.Body.Close()

	drugs, err := ParseDrugs(res.Body)
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("parse drugs: %w", err)
	}

	return drugs, nil
}

// ParseDrugs parses the drugs from the given reader.
func ParseDrugs(r io.Reader) (DrugsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return DrugsResponse{}, fmt.Errorf("parse HTML: %w", err)
	}

	var drugs []Drug
	doc.Find("tr[data-key]").Each(func(i int, s *goquery.Selection) {
		drug, err := parseDrug(s)
		if err != nil {
			log.Printf("error parsing drug #%d: %s", i, err)
			return
		}

		drugs = append(drugs, drug)
	})

	return DrugsResponse{Drugs: drugs, OtherPages: parsePagination(doc)}, nil
}

func parseDrug(selection *goquery.Selection) (Drug, error) {
	var drug Drug
	if err := UnmarshalDataColumnByIndex("drugs-index", selection, &drug); err != nil {
		return Drug{}, fmt.Errorf("unmarshal drug: %w", err)
	}

	// Get the value from <a class="pilih" value="123" href="/obat-batch/index" title="Detail">some image</a>
	idStr, ok := selection.Find("a.pilih").Attr("value")
	if !ok {
		return Drug{}, fmt.Errorf("drug vmedis id not found")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return Drug{}, fmt.Errorf("parse drug vmedis id: %w", err)
	}

	drug.VmedisID = id

	return drug, nil
}
