package drug

import (
	"fmt"
	"slices"
	"strings"

	"github.com/turfaa/vmedis-proxy-api/auth"

	"github.com/gin-gonic/gin"
	"github.com/leekchan/accounting"
)

var (
	rupiah = accounting.Accounting{Symbol: "Rp", Format: "%s %v", FormatZero: "%s 0", Thousand: ".", Decimal: ","}
)

// GetDrugsV2 handles row-based get drugs request.
func (h *ApiHandler) GetDrugsV2(c *gin.Context) {
	user := auth.FromGinContext(c)

	drugs, err := h.service.GetDrugs(c)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("failed to get drugs: %s", err),
		})
		return
	}

	res := DrugsResponseV2{
		Drugs: h.transformToDrugsV2(user, drugs),
	}

	c.JSON(200, res)
}

func (h *ApiHandler) transformToDrugsV2(user auth.User, drugs []Drug) []DrugsResponseV2_Drug {
	transformedDrugs := make([]DrugsResponseV2_Drug, len(drugs))
	for i, drug := range drugs {
		transformedDrugs[i] = h.transformToDrugV2(user, drug)
	}

	return transformedDrugs
}

func (h *ApiHandler) transformToDrugV2(user auth.User, drug Drug) DrugsResponseV2_Drug {
	sections := make([]Section, 0, 5)
	addSection := func(allowedRoles []auth.Role, title string, rowBuilder func() []string) {
		if slices.Contains(allowedRoles, user.Role) {
			sections = append(sections, Section{
				Title: title,
				Rows:  rowBuilder(),
			})
		}
	}

	units := filterUnits(drug.Units)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff, auth.RoleReseller, auth.RoleGuest},
		"Harga Normal",
		func() []string {
			rows := make([]string, len(units))
			for i, unit := range units {
				rows[i] = fmt.Sprintf("%s / %s", rupiah.FormatMoney(unit.PriceOne), unit.Unit)
				if unit.ConversionToParentUnit > 0 {
					rows[i] += fmt.Sprintf(" (%0.0f %s)", unit.ConversionToParentUnit, unit.ParentUnit)
				}
			}

			return rows
		},
	)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff, auth.RoleReseller},
		"Harga Diskon",
		func() []string {
			rows := make([]string, len(units))
			for i, unit := range units {
				rows[i] = fmt.Sprintf("%s / %s", rupiah.FormatMoney(unit.PriceTwo), unit.Unit)
				if unit.ConversionToParentUnit > 0 {
					rows[i] += fmt.Sprintf(" (%0.0f %s)", unit.ConversionToParentUnit, unit.ParentUnit)
				}
			}

			return rows
		},
	)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff},
		"Harga Resep",
		func() []string {
			rows := make([]string, len(units))
			for i, unit := range units {
				rows[i] = fmt.Sprintf("%s / %s", rupiah.FormatMoney(unit.PriceThree), unit.Unit)
				if unit.ConversionToParentUnit > 0 {
					rows[i] += fmt.Sprintf(" (%0.0f %s)", unit.ConversionToParentUnit, unit.ParentUnit)
				}
			}

			return rows
		},
	)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff, auth.RoleReseller, auth.RoleGuest},
		"Sisa Stok",
		func() []string {
			unitStrings := make([]string, len(drug.Stocks))
			for i, stock := range drug.Stocks {
				unitStrings[i] = stock.String()
			}

			return []string{strings.Join(unitStrings, " ")}
		},
	)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff},
		"Stok Minimum",
		func() []string {
			return []string{drug.MinimumStock.String()}
		},
	)

	addSection(
		[]auth.Role{auth.RoleAdmin, auth.RoleStaff},
		"Kode Obat Vmedis",
		func() []string {
			return []string{drug.VmedisCode}
		},
	)

	return DrugsResponseV2_Drug{
		VmedisCode: drug.VmedisCode,
		Name:       drug.Name,
		Sections:   sections,
	}
}
