package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type DashboardHandler struct {
	dashboard *service.DashboardService
}

func NewDashboardHandler(dashboard *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

type DashboardSummaryResponse struct {
	Totals      DashboardTotalsResponse       `json:"totals"`
	Aging       DashboardAgingResponse        `json:"aging"`
	Brokers     []DashboardBrokerStatResponse `json:"brokers"`
	Dealers     []DashboardDealerStatResponse `json:"dealers"`
	RecentDeals []DealResponse                `json:"recent_deals"`
}

type DashboardTotalsResponse struct {
	TotalDeals             int64    `json:"total_deals"`
	GrossRevenue           *float64 `json:"gross_revenue"`
	BrokerPayout           *float64 `json:"broker_payout"`
	TotalDocFees           *float64 `json:"total_doc_fees"`
	CompanyProfitBeforeTax *float64 `json:"company_profit_before_tax"`
	PaidRevenue            *float64 `json:"paid_revenue"`
	PendingRevenue         *float64 `json:"pending_revenue"`
	PaidBrokerPayout       *float64 `json:"paid_broker_payout"`
	UnpaidBrokerPayout     *float64 `json:"unpaid_broker_payout"`
	PendingCompanyProfit   *float64 `json:"pending_company_profit"`
	AvgRevenuePerDeal      *float64 `json:"avg_revenue_per_deal"`
	AvgProfitPerDeal       *float64 `json:"avg_profit_per_deal"`
	TotalReferralPayout    *float64 `json:"total_referral_payout"`
}

type DashboardAgingResponse struct {
	Pending031Count      int64    `json:"pending_0_31_count"`
	Pending031Revenue    *float64 `json:"pending_0_31_revenue"`
	Pending3240Count     int64    `json:"pending_32_40_count"`
	Pending3240Revenue   *float64 `json:"pending_32_40_revenue"`
	Pending41PlusCount   int64    `json:"pending_41_plus_count"`
	Pending41PlusRevenue *float64 `json:"pending_41_plus_revenue"`
	OldestPendingDays    int32    `json:"oldest_pending_days"`
}

type DashboardBrokerStatResponse struct {
	ID                     string   `json:"id"`
	FirstName              string   `json:"first_name"`
	LastName               string   `json:"last_name"`
	DealCount              int64    `json:"deal_count"`
	GrossRevenue           *float64 `json:"gross_revenue"`
	BrokerPayout           *float64 `json:"broker_payout"`
	CompanyProfitBeforeTax *float64 `json:"company_profit_before_tax"`
	AvgRevenuePerDeal      *float64 `json:"avg_revenue_per_deal"`
}

type DashboardDealerStatResponse struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	DealCount        int64    `json:"deal_count"`
	GrossRevenue     *float64 `json:"gross_revenue"`
	AvgDaysToPayment *float64 `json:"avg_days_to_payment"`
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	fromDate, err := optionalQueryDate(r, "from")
	if err != nil {
		http.Error(w, "invalid from date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	toDate, err := optionalQueryDate(r, "to")
	if err != nil {
		http.Error(w, "invalid to date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	if fromDate != nil && toDate != nil && fromDate.After(*toDate) {
		http.Error(w, "from date must be before or equal to to date", http.StatusBadRequest)
		return
	}

	limit := int32(10)
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = int32(n)
	}

	summary, err := h.dashboard.Summary(r.Context(), service.DashboardInput{
		FromDate: fromDate,
		ToDate:   toDate,
		Limit:    limit,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDashboardSummaryResponse(summary))
}

func optionalQueryDate(r *http.Request, key string) (*time.Time, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func toDashboardSummaryResponse(s service.DashboardSummary) DashboardSummaryResponse {
	return DashboardSummaryResponse{
		Totals:      toDashboardTotalsResponse(s.Totals),
		Aging:       toDashboardAgingResponse(s.Aging),
		Brokers:     toDashboardBrokerStatResponses(s.Brokers),
		Dealers:     toDashboardDealerStatResponses(s.Dealers),
		RecentDeals: toDealResponses(s.RecentDeals),
	}
}

func toDashboardTotalsResponse(t db.GetDashboardTotalsRow) DashboardTotalsResponse {
	return DashboardTotalsResponse{
		TotalDeals:             t.TotalDeals,
		GrossRevenue:           pgNumeric(t.GrossRevenue),
		BrokerPayout:           pgNumeric(t.BrokerPayout),
		TotalDocFees:           pgNumeric(t.TotalDocFees),
		CompanyProfitBeforeTax: pgNumeric(t.CompanyProfitBeforeTax),
		PaidRevenue:            pgNumeric(t.PaidRevenue),
		PendingRevenue:         pgNumeric(t.PendingRevenue),
		PaidBrokerPayout:       pgNumeric(t.PaidBrokerPayout),
		UnpaidBrokerPayout:     pgNumeric(t.UnpaidBrokerPayout),
		PendingCompanyProfit:   pgNumeric(t.PendingCompanyProfit),
		AvgRevenuePerDeal:      pgNumeric(t.AvgRevenuePerDeal),
		AvgProfitPerDeal:       pgNumeric(t.AvgProfitPerDeal),
		TotalReferralPayout:    pgNumeric(t.TotalReferralPayout),
	}
}

func toDashboardAgingResponse(a db.GetDashboardAgingRow) DashboardAgingResponse {
	return DashboardAgingResponse{
		Pending031Count:      a.Pending031Count,
		Pending031Revenue:    pgNumeric(a.Pending031Revenue),
		Pending3240Count:     a.Pending3240Count,
		Pending3240Revenue:   pgNumeric(a.Pending3240Revenue),
		Pending41PlusCount:   a.Pending41PlusCount,
		Pending41PlusRevenue: pgNumeric(a.Pending41PlusRevenue),
		OldestPendingDays:    a.OldestPendingDays,
	}
}

func toDashboardBrokerStatResponses(rows []db.ListDashboardBrokerStatsRow) []DashboardBrokerStatResponse {
	out := make([]DashboardBrokerStatResponse, len(rows))
	for i, row := range rows {
		out[i] = DashboardBrokerStatResponse{
			ID:                     pgUUID(row.ID),
			FirstName:              row.FirstName,
			LastName:               row.LastName,
			DealCount:              row.DealCount,
			GrossRevenue:           pgNumeric(row.GrossRevenue),
			BrokerPayout:           pgNumeric(row.BrokerPayout),
			CompanyProfitBeforeTax: pgNumeric(row.CompanyProfitBeforeTax),
			AvgRevenuePerDeal:      pgNumeric(row.AvgRevenuePerDeal),
		}
	}
	return out
}

func toDashboardDealerStatResponses(rows []db.ListDashboardDealerStatsRow) []DashboardDealerStatResponse {
	out := make([]DashboardDealerStatResponse, len(rows))
	for i, row := range rows {
		out[i] = DashboardDealerStatResponse{
			ID:               pgUUID(row.ID),
			Name:             row.Name,
			DealCount:        row.DealCount,
			GrossRevenue:     pgNumeric(row.GrossRevenue),
			AvgDaysToPayment: pgNumeric(row.AvgDaysToPayment),
		}
	}
	return out
}
