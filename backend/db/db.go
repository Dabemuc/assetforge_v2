package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	db   *sql.DB
	once sync.Once
)

func Establish_db_conn() {
	once.Do(func() {
		// Load .env file according to app env
		var app_env = os.Getenv("APP_ENV")
		if app_env == "" {
			app_env = "dev"
		}
		_, filename, _, _ := runtime.Caller(0)                           // Gets this files path
		projectRoot := filepath.Join(filepath.Dir(filename), "..", "..") // Goes up two levels to root dir
		envFile := filepath.Join(projectRoot, "backend", app_env+".env")
		err := godotenv.Load(envFile)
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}

		// Read env vars
		DB_USER := os.Getenv("ASSETFORGE_V2_DB_USER")
		DB_PASSWORD := os.Getenv("ASSETFORGE_V2_DB_PASSWORD")
		DB_NAME := os.Getenv("ASSETFORGE_V2_DB_NAME")
		DB_HOST := os.Getenv("ASSETFORGE_V2_DB_HOST")
		DB_PORT := os.Getenv("ASSETFORGE_V2_DB_PORT")

		// Connect to db
		connStr := "user=" + DB_USER + " password=" + DB_PASSWORD + " dbname=" + DB_NAME + " sslmode=disable host=" + DB_HOST + " port=" + DB_PORT
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal("Error connecting to the database:", err)
		}
		err = db.Ping()
		if err != nil {
			log.Fatal("Connection could not be opened. Error:", err)
		}
		log.Print("Successfully conected to database!")
	})
}

func GetDb() *sql.DB {
	if db == nil {
		log.Fatal("db.Getdb() called without connection being established first. First call db.Establish_db_conn().")
	}
	return db
}

func InsertOrUpdateEtf(id string, name string, fundVolume string, isDistributing bool, releaseDate time.Time, replicationMethod string, shareClassVolume string, totalExpenseRatio float32) {
	// Ensure releaseDate is only a date, not a timestamp.
	releaseDate = releaseDate.Truncate(24 * time.Hour)

	scrape_date_base_data := time.Now()

	var queryString = `
		INSERT INTO t_etf (id, name, fundVolume, isDistributing, releaseDate, replicationMethod, shareClassVolume, totalExpenseRatio, scrape_date_base_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id)
		DO UPDATE SET
			name = EXCLUDED.name,
			fundVolume = EXCLUDED.fundVolume,
			isDistributing = EXCLUDED.isDistributing,
			releaseDate = EXCLUDED.releaseDate,
			replicationMethod = EXCLUDED.replicationMethod,
			shareClassVolume = EXCLUDED.shareClassVolume,
			totalExpenseRatio = EXCLUDED.totalExpenseRatio,
      scrape_date_base_data = EXCLUDED.scrape_date_base_data`

	_, err := db.Exec(queryString, id, name, fundVolume, isDistributing, releaseDate, replicationMethod, shareClassVolume, totalExpenseRatio, scrape_date_base_data)
	if err != nil {
		log.Printf("Error executing query: %v", err)
	}
}

type EtfDetailsData struct {
	Id                         string
	ISIN                       string `json:"isin"`
	WKN                        string `json:"wkn"`
	NrPositions                string `json:"nr_positions"`
	BaseIndex                  string `json:"base_index"`
	ShareClassVolume           string `json:"share_class_volume"`
	FundDomicile               string `json:"fund_domicile"`
	FundCurrency               string `json:"fund_currency"`
	SecuritiesLendingPermitted bool   `json:"securities_lending_permitted"`
	TradeCurrency              string `json:"trade_currency"`
	HasCurrencyHedging         bool   `json:"has_currency_hedging"`
	HasSpecialAssets           bool   `json:"has_special_assets"`
	FundProvider               string `json:"fund_provider"`
	LegalStructure             string `json:"legal_structure"`
	FundStructure              string `json:"fund_structure"`
	Administrator              string `json:"administrator"`
	Depotbank                  string `json:"depotbank"`
	Auditor                    string `json:"auditor"`

	CountryComposition []struct {
		Country    string `json:"country"`
		Percentile string `json:"percentile"`
	} `json:"country_composition"`

	RegionComposition []struct {
		Country    string `json:"country"`
		Percentile string `json:"percentile"`
	} `json:"region_composition"`

	CurrencyDistribution []struct {
		Country    string `json:"country"`
		Percentile string `json:"percentile"`
	} `json:"currency_distribution"`

	WeightTop10             string `json:"weight_top_10"`
	NrStockPositions        string `json:"nr_stock_positions"`
	NrBondPositions         string `json:"nr_bond_positions"`
	NrCashAndOtherPositions string `json:"nr_cash_and_other_positions"`

	Top10Holdings []struct {
		Name       string `json:"name"`
		Percentile string `json:"percentile"`
	} `json:"top_10_holdings"`

	IndustryDistribution []struct {
		Name       string `json:"name"`
		Percentile string `json:"percentile"`
	} `json:"industry_distribution"`

	ActivityDistribution []struct {
		Name        string `json:"name"`
		Percentiles struct {
			Min   string `json:"min"`
			Value string `json:"value"`
			Max   string `json:"max"`
		} `json:"percentiles"`
	} `json:"activity_distribution"`

	HistoricalPerformance []struct {
		Timespan    string `json:"timespan"`
		Performance string `json:"performance"`
		Return      string `json:"return"`
	} `json:"historical_performance"`

	HistoricalVolatility []struct {
		Period string `json:"period"`
		Value  string `json:"value"`
	} `json:"historical_volatility"`

	HistoricalMaxDrawdown []struct {
		Period string `json:"period"`
		Value  string `json:"value"`
	} `json:"historical_max_drawdown"`

	HistoricalSharpeRatio []struct {
		Period string `json:"period"`
		Value  string `json:"value"`
	} `json:"historical_sharpe_ratio"`

	Exchanges []struct {
		Name     string `json:"name"`
		Currency string `json:"currency"`
		Ticker   string `json:"ticker"`
	} `json:"exchanges"`
}

func UpdateEtfDetails(data EtfDetailsData) error {
	var scrapeDateDetails = time.Now()

	// parse
	var weight_top_10 = strings.TrimSpace(data.WeightTop10)
	weight_top_10 = strings.TrimSuffix(weight_top_10, "%")
	weight_top_10 = strings.ReplaceAll(weight_top_10, "\u00a0", "")
	weight_top_10 = strings.ReplaceAll(weight_top_10, ",", ".")
	var weight_top_10_float, err_weight_top_10_float = strconv.ParseFloat(weight_top_10, 32)
	weight_top_10_float = weight_top_10_float / 100
	if err_weight_top_10_float != nil {
		fmt.Println("Error parsing totalExpenseRatio:", err_weight_top_10_float)
	}

	// Convert JSON fields to string
	countryComposition, err := json.Marshal(data.CountryComposition)
	if err != nil {
		return err
	}
	regionComposition, err := json.Marshal(data.RegionComposition)
	if err != nil {
		return err
	}
	currencyDistribution, err := json.Marshal(data.CurrencyDistribution)
	if err != nil {
		return err
	}
	top10Holdings, err := json.Marshal(data.Top10Holdings)
	if err != nil {
		return err
	}
	industryDistribution, err := json.Marshal(data.IndustryDistribution)
	if err != nil {
		return err
	}
	activityDistribution, err := json.Marshal(data.ActivityDistribution)
	if err != nil {
		return err
	}
	historicalPerformance, err := json.Marshal(data.HistoricalPerformance)
	if err != nil {
		return err
	}
	historicalVolatility, err := json.Marshal(data.HistoricalVolatility)
	if err != nil {
		return err
	}
	historicalMaxDrawdown, err := json.Marshal(data.HistoricalMaxDrawdown)
	if err != nil {
		return err
	}
	historicalSharpeRatio, err := json.Marshal(data.HistoricalSharpeRatio)
	if err != nil {
		return err
	}
	exchanges, err := json.Marshal(data.Exchanges)
	if err != nil {
		return err
	}

	query := `
		UPDATE t_etf SET
			isin = $2,
			wkn = $3,
			nr_positions = $4,
			base_index = $5,
			share_class_volume = $6,
			fund_domicile = $7,
			fund_currency = $8,
			securities_lending_permitted = $9,
			trade_currency = $10,
			has_currency_hedging = $11,
			has_special_assets = $12,
			fund_provider = $13,
			legal_structure = $14,
			fund_structure = $15,
			administrator = $16,
			depotbank = $17,
			auditor = $18,
			country_composition = $19,
			region_composition = $20,
			currency_distribution = $21,
			weight_top_10 = $22,
			nr_stock_positions = $23,
			nr_bond_positions = $24,
			nr_cash_and_other_positions = $25,
			top_10_holdings = $26,
			industry_distribution = $27,
			activity_distribution = $28,
			historical_performance = $29,
			historical_volatility = $30,
			historical_max_drawdown = $31,
			historical_sharpe_ratio = $32,
			exchanges = $33,
			scrape_date_details = $34
		WHERE id = $1
	`

	_, err = db.Exec(query, data.Id, data.ISIN, data.WKN, data.NrPositions, data.BaseIndex, data.ShareClassVolume,
		data.FundDomicile, data.FundCurrency, data.SecuritiesLendingPermitted, data.TradeCurrency,
		data.HasCurrencyHedging, data.HasSpecialAssets, data.FundProvider, data.LegalStructure,
		data.FundStructure, data.Administrator, data.Depotbank, data.Auditor, countryComposition,
		regionComposition, currencyDistribution, weight_top_10_float, data.NrStockPositions,
		data.NrBondPositions, data.NrCashAndOtherPositions, top10Holdings, industryDistribution,
		activityDistribution, historicalPerformance, historicalVolatility, historicalMaxDrawdown,
		historicalSharpeRatio, exchanges, scrapeDateDetails)
	if err != nil {
		log.Printf("Error executing update: %v", err)
		return err
	}

	return nil
}
