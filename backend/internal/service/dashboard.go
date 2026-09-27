package service

import (
	"context"
	"time"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type DashboardService struct {
	q *db.Queries
}

func NewDashboard(q *db.Queries) *DashboardService {
	return &DashboardService{q: q}
}

type DashboardInput struct {
	FromDate    *time.Time
	ToDate      *time.Time
	Limit       int32
	ExcludeCash bool
}

type DashboardSummary struct {
	Totals      db.GetDashboardTotalsRow
	Aging       db.GetDashboardAgingRow
	Brokers     []db.ListDashboardBrokerStatsRow
	Dealers     []db.ListDashboardDealerStatsRow
	RecentDeals []db.Deal
	Leases      []db.ListDashboardLeasesRow
}

func (s *DashboardService) Summary(ctx context.Context, in DashboardInput) (DashboardSummary, error) {
	if in.Limit <= 0 {
		in.Limit = 10
	}
	if in.Limit > 50 {
		in.Limit = 50
	}

	fromDate := toDate(in.FromDate)
	toDate := toDate(in.ToDate)
	excludeCash := toBool(in.ExcludeCash)

	totals, err := s.q.GetDashboardTotals(ctx, db.GetDashboardTotalsParams{
		FromDate:    fromDate,
		ToDate:      toDate,
		ExcludeCash: excludeCash,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	aging, err := s.q.GetDashboardAging(ctx, db.GetDashboardAgingParams{
		FromDate:    fromDate,
		ToDate:      toDate,
		ExcludeCash: excludeCash,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	brokers, err := s.q.ListDashboardBrokerStats(ctx, db.ListDashboardBrokerStatsParams{
		FromDate: fromDate,
		ToDate:   toDate,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	dealers, err := s.q.ListDashboardDealerStats(ctx, db.ListDashboardDealerStatsParams{
		FromDate: fromDate,
		ToDate:   toDate,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	recentDeals, err := s.q.ListDashboardRecentDeals(ctx, db.ListDashboardRecentDealsParams{
		Limit:    in.Limit,
		FromDate: fromDate,
		ToDate:   toDate,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	leases, err := s.q.ListDashboardLeases(ctx, db.ListDashboardLeasesParams{
		FromDate:    fromDate,
		ToDate:      toDate,
		ExcludeCash: excludeCash,
	})
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		Totals:      totals,
		Aging:       aging,
		Brokers:     brokers,
		Dealers:     dealers,
		RecentDeals: recentDeals,
		Leases:      leases,
	}, nil
}
