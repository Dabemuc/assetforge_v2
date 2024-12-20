-- Migration Down

ALTER TABLE IF EXISTS t_etf 
DROP COLUMN IF EXISTS scrape_date_base_data,
DROP COLUMN IF EXISTS scrape_date_details,
DROP COLUMN IF EXISTS isin,
DROP COLUMN IF EXISTS wkn,
DROP COLUMN IF EXISTS nr_positions,
DROP COLUMN IF EXISTS base_index,
DROP COLUMN IF EXISTS share_class_volume,
DROP COLUMN IF EXISTS fund_domicile,
DROP COLUMN IF EXISTS fund_currency,
DROP COLUMN IF EXISTS securities_lending_permitted,
DROP COLUMN IF EXISTS trade_currency,
DROP COLUMN IF EXISTS has_currency_hedging,
DROP COLUMN IF EXISTS has_special_assets,
DROP COLUMN IF EXISTS fund_provider,
DROP COLUMN IF EXISTS legal_structure,
DROP COLUMN IF EXISTS fund_structure,
DROP COLUMN IF EXISTS administrator,
DROP COLUMN IF EXISTS depotbank,
DROP COLUMN IF EXISTS auditor,
DROP COLUMN IF EXISTS country_composition,
DROP COLUMN IF EXISTS region_composition,
DROP COLUMN IF EXISTS currency_distribution,
DROP COLUMN IF EXISTS weight_top_10,
DROP COLUMN IF EXISTS nr_stock_positions,
DROP COLUMN IF EXISTS nr_bond_positions,
DROP COLUMN IF EXISTS nr_cash_and_other_positions,
DROP COLUMN IF EXISTS top_10_holdings,
DROP COLUMN IF EXISTS industry_distribution,
DROP COLUMN IF EXISTS activity_distribution,
DROP COLUMN IF EXISTS historical_performance,
DROP COLUMN IF EXISTS historical_volatility,
DROP COLUMN IF EXISTS historical_max_drawdown,
DROP COLUMN IF EXISTS historical_sharpe_ratio,
DROP COLUMN IF EXISTS exchanges;
