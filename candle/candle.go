package candle

// https://mholt.github.io/json-to-go/
type Ohlcv []struct {
  Datetime int64   `json:"datetime"`
  Open     float64 `json:"open"`
  High     float64 `json:"high"`
  Low      float64 `json:"low"`
  Close    float64 `json:"close"`
  Volume   float64 `json:"volume"`
  Rsi4H    float64 `json:"rsi_4h"`
  EmaRsi4H float64 `json:"ema_rsi_4h"`
  Ema84H   float64 `json:"ema_8_4h"`
  Ema144H  float64 `json:"ema_14_4h"`
  Ema504H  float64 `json:"ema_50_4h"`
  Fastd4H  float64 `json:"fastd_4h"`
  Fastk4H  float64 `json:"fastk_4h"`
  Macd4H   float64 `json:"macd_4h"`
  Macds4H  float64 `json:"macds_4h"`
  Macdh4H  float64 `json:"macdh_4h"`
  Rsi1D    float64 `json:"rsi_1d"`
  Rsi1W    float64 `json:"rsi_1w"`
  EmaRsi1D float64 `json:"ema_rsi_1d"`
  EmaRsi1W float64 `json:"ema_rsi_1w"`
  Fastd1D  float64 `json:"fastd_1d"`
  Fastk1D  float64 `json:"fastk_1d"`
  Fastd1W  float64 `json:"fastd_1w"`
  Fastk1W  float64 `json:"fastk_1w"`
  Macd1D   float64 `json:"macd_1d"`
  Macds1D  float64 `json:"macds_1d"`
  Macdh1D  float64 `json:"macdh_1d"`
  Macd1W   float64 `json:"macd_1w"`
  Macds1W  float64 `json:"macds_1w"`
  Macdh1W  float64 `json:"macdh_1w"`
  Ema81D   float64 `json:"ema_8_1d"`
  Ema81W   float64 `json:"ema_8_1w"`
  Ema141D  float64 `json:"ema_14_1d"`
  Ema141W  float64 `json:"ema_14_1w"`
  Ema501D  float64 `json:"ema_50_1d"`
  Ema501W  float64 `json:"ema_50_1w"`
}
