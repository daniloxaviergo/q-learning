package trade

import (
	"q-learning/candle"
	"sort"
	"strconv"
	"sync"
)

type Trade struct {}

func hasHigher(open float64, _ int, idx int, myPrice float64, priceWithProfit float64, ch chan float64, wg *sync.WaitGroup) {
  defer wg.Done()

  var result float64
  result = -1.0
  maxSteps := 31
  stepsToSell := idx + 1

  // stepsToSell := (step - idx) + 1
  // fmt.Println("step idx", step, idx)
  // fmt.Println("stepsToSell", stepsToSell)
  if(stepsToSell >= maxSteps) {
    ch <- result
    return
  }

  sellAmount := myPrice * open
  if(sellAmount >= priceWithProfit) {
    // punishment := float64(stepsToSell) * (1.0 / float64(maxSteps))
    punishment := float64(stepsToSell) * 0.023
    // fmt.Println("punishment", punishment)
    result = 1.0 - punishment
  }

  ch <- result
}

func (trade *Trade) Reward(current_step string, ohlcvs candle.Ohlcv, wallet float64, roi float64) float64 {
  var wg sync.WaitGroup
  // step, _ := strconv.ParseInt(current_step, 10, 32)
  step, _ := strconv.Atoi(current_step)

  // fazer loop dividindo 15 + 15
  // assim teremos dois channels processando em paralelo
  // se o primeiro loop retornar valor não é necessário fazer outro
  currentPrice := ohlcvs[step].Open
  // fmt.Println(ohlcvs[step].Open)
  // fmt.Println(ohlcvs[step].Datetime)
  // agendar goroutine somente quem tem o idx <= 80
  lenOhlcvs := len(ohlcvs)
  var endOhlcvs int
  if((step+30) >= lenOhlcvs) {
    endOhlcvs = lenOhlcvs
  } else {
    endOhlcvs = step + 30
  }
  limited_ohlcvs := ohlcvs[(step+1):(endOhlcvs)]

  // var wallet float64
  // wallet = 4000.00

  // 0.003 => 0.3% # only fees
  // 0.013 => 1.3%
  // 0.033 => 3.3%
  // 0.103 => 10.3%
  // 0.303 => 30.3%
  priceWithProfit := (wallet * roi) + wallet
  myPrice := wallet / currentPrice

  // fmt.Println(currentPrice)
  ch := make(chan float64, len(limited_ohlcvs))
  for idx, ohlcv := range limited_ohlcvs {
    wg.Add(1)
    go hasHigher(ohlcv.Open, step, idx, myPrice, priceWithProfit, ch, &wg)
  }

  go func() {
    wg.Wait()
    close(ch)
  }()

  resp := -1.0
  var responses []float64

  for result := range ch {
    if(result > 0) {
      responses = append(responses, result)
    }
  }

  if(len(responses) > 0) {
    sort.Float64s(responses)
    // minimum := responses[0]
    // maximum := responses[len(responses)-1]
    // fmt.Println(minimum)
    // fmt.Println(maximum)
    // fmt.Println(responses)
    resp = responses[len(responses) - 1]
  }

  // fmt.Println(resp)
  return resp
}
