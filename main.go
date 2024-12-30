package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"q-learning/brain"
	"q-learning/candle"
	"q-learning/state"
	"q-learning/trade"
	"strconv"
	"strings"
	"github.com/redis/go-redis/v9"
)

// fazer server http
// implementar load do json dos candles
// implementar q-table(structs, get_value, learning, etc...)
// salvar/carregar no redis

func loadCurrency(candles **candle.Ohlcv, currency string) {
  jsonFile, err := os.Open("/home/danilo/scripts/forecast/tmp/" + currency + ".json")
  if err != nil {
    fmt.Println(err)
  }

  byteValue, _ := ioutil.ReadAll(jsonFile)
  // var candles *candle.Ohlcv

  json.Unmarshal(byteValue, &candles)
  jsonFile.Close()

  // real_ohlcvs := *candles

  // for i := 0; i < len(real_ohlcvs); i++ {
  //   fmt.Println("Open: %f", real_ohlcvs[i].Open)
  // }
}

func loadQTable(brain *brain.Brain) {
  ctx := context.Background()
  client := redis.NewClient(&redis.Options{
    Addr:     "freqai.pro:6379",
    Password: "", // no password set
    DB:     3,  // use default DB
  })

  keys, kError := client.Keys(ctx, "brain:*").Result()
  // fmt.Println(a)
  // fmt.Println(b)
  // a := client.Get(ctx, "brain:1111")
  fmt.Println("Load from redis", len(keys))
  if(kError != nil) {
    fmt.Println("Redis not running....")
    os.Exit(1)
  }
  // if(len(keys) < 10) {
  //   fmt.Println("Redis not running....")
  //   os.Exit(1)
  // }

  for _, key := range keys {
    value, _ := client.Get(ctx, key).Result()

    redis_key := strings.Split(key, ":")
    redis_struct_values := strings.Split(redis_key[1], "|")
    redis_values := strings.Split(value, "|")

    var state state.State
    state.Rsi4H    = redis_struct_values[0]
    state.EmaRsi4H = redis_struct_values[1]
    state.Ema84H   = redis_struct_values[2]
    state.Ema144H  = redis_struct_values[3]
    state.Ema504H  = redis_struct_values[4]
    state.Fastd4H  = redis_struct_values[5]
    state.Fastk4H  = redis_struct_values[6]
    state.Macd4H   = redis_struct_values[7]
    state.Macds4H  = redis_struct_values[8]
    state.Macdh4H  = redis_struct_values[9]
    state.Rsi1D    = redis_struct_values[10]
    state.Rsi1W    = redis_struct_values[11]
    state.EmaRsi1D = redis_struct_values[12]
    state.EmaRsi1W = redis_struct_values[13]
    state.Fastd1D  = redis_struct_values[14]
    state.Fastk1D  = redis_struct_values[15]
    state.Fastd1W  = redis_struct_values[16]
    state.Fastk1W  = redis_struct_values[17]
    state.Macd1D   = redis_struct_values[18]
    state.Macds1D  = redis_struct_values[19]
    state.Macdh1D  = redis_struct_values[20]
    state.Macd1W   = redis_struct_values[21]
    state.Macds1W  = redis_struct_values[22]
    state.Macdh1W  = redis_struct_values[23]
    state.Ema81D   = redis_struct_values[24]
    state.Ema81W   = redis_struct_values[25]
    state.Ema141D  = redis_struct_values[26]
    state.Ema141W  = redis_struct_values[27]
    state.Ema501D  = redis_struct_values[28]
    state.Ema501W  = redis_struct_values[29]
    state.Label    = redis_struct_values[30]

    // fmt.Println(redis_struct_values)
    // fmt.Println(redis_values[0])
    // fmt.Println(redis_values[1])
    // fmt.Println(state)

    actions := [2]float64{0.0, 0.0}
    // brain.QTable[state] = [2]float64{1.0, 20.0}
    // value, _ := strconv.ParseFloat(strings.TrimSpace(redis_values[0]), 64)
    f, err := strconv.ParseFloat(strings.TrimSpace(redis_values[0]), 64)
    if err != nil {
      panic(err)
    } else {
      actions[0] = f
    }

    f1, err := strconv.ParseFloat(strings.TrimSpace(redis_values[1]), 64)
    if err != nil {
      panic(err)
    } else {
      actions[1] = f1
    }

    brain.QTable[state] = actions
    // brain.QTable[state][0] = value
    // brain.QTable[state][1], _ := strconv.ParseFloat(strings.TrimSpace(redis_values[1]), 64)
  }
}

func main() {
  args := os.Args[1:]
  port := args[0]

  var stateKeyLock = state.NewStateKeyLock()
  var brain brain.Brain
  brain.QTable = make(map[state.State][2]float64)

  loadQTable(&brain)
  brain.Label = "1"
  brain.Wallet = 4000.0
  // 0.003 => 0.3% # only fees
  // 0.013 => 1.3%
  // 0.033 => 3.3%
  // 0.103 => 10.3%
  // 0.303 => 30.3%
  brain.Roi = 0.303

  // Start TCP server
  listener, err := net.Listen("tcp", ":" + port)
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }
  defer listener.Close()
  fmt.Println("Server listening on port ", port)

  var candles *candle.Ohlcv
  loadCurrency(&candles, "sol")

  for {
    conn, err := listener.Accept()
    if err != nil {
      log.Printf("Error accepting connection: %v", err)
      continue
    }
    
    go handleConnection(conn, candles, stateKeyLock, &brain)
  }
}

func handleConnection(conn net.Conn, candles *candle.Ohlcv, stateKeyLock *state.StateKeyLock, brain *brain.Brain) {
  defer conn.Close()

  scanner := bufio.NewScanner(conn)
  for scanner.Scan() {
    request := strings.Fields(scanner.Text())
    if len(request) == 0 {
      continue
    }

    switch strings.ToUpper(request[0]) {
    case "LOAD_CURRENCY":
      if len(request) != 2 {
        conn.Write([]byte("Invalid command load currency: set currency\n"))
        continue
      }

      currency := request[1]
      loadCurrency(&candles, currency)

      conn.Write([]byte(fmt.Sprintf("ok\n")))
    case "REWARD":
      if len(request) != 2 {
        conn.Write([]byte("Invalid command load currency: set currency\n"))
        continue
      }

      var trade trade.Trade
      current_step := request[1]
      reward := trade.Reward(current_step, *candles, brain.Wallet, brain.Roi)

      conn.Write([]byte(fmt.Sprintf("%s\n", strconv.FormatFloat(reward, 'f', -1, 64))))
    case "LEARNING":
      currency := request[1]

      var current_state state.State
      current_state.Rsi4H    = request[2]
      current_state.EmaRsi4H = request[3]
      current_state.Ema84H   = request[4]
      current_state.Ema144H  = request[5]
      current_state.Ema504H  = request[6]
      current_state.Fastd4H  = request[7]
      current_state.Fastk4H  = request[8]
      current_state.Macd4H   = request[9]
      current_state.Macds4H  = request[10]
      current_state.Macdh4H  = request[11]
      current_state.Rsi1D    = request[12]
      current_state.Rsi1W    = request[13]
      current_state.EmaRsi1D = request[14]
      current_state.EmaRsi1W = request[15]
      current_state.Fastd1D  = request[16]
      current_state.Fastk1D  = request[17]
      current_state.Fastd1W  = request[18]
      current_state.Fastk1W  = request[19]
      current_state.Macd1D   = request[20]
      current_state.Macds1D  = request[21]
      current_state.Macdh1D  = request[22]
      current_state.Macd1W   = request[23]
      current_state.Macds1W  = request[24]
      current_state.Macdh1W  = request[25]
      current_state.Ema81D   = request[26]
      current_state.Ema81W   = request[27]
      current_state.Ema141D  = request[28]
      current_state.Ema141W  = request[29]
      current_state.Ema501D  = request[30]
      current_state.Ema501W  = request[31]
      current_state.Label    = brain.Label

      var next_state state.State
      next_state.Rsi4H    = request[32]
      next_state.EmaRsi4H = request[33]
      next_state.Ema84H   = request[34]
      next_state.Ema144H  = request[35]
      next_state.Ema504H  = request[36]
      next_state.Fastd4H  = request[37]
      next_state.Fastk4H  = request[38]
      next_state.Macd4H   = request[39]
      next_state.Macds4H  = request[40]
      next_state.Macdh4H  = request[41]
      next_state.Rsi1D    = request[42]
      next_state.Rsi1W    = request[43]
      next_state.EmaRsi1D = request[44]
      next_state.EmaRsi1W = request[45]
      next_state.Fastd1D  = request[46]
      next_state.Fastk1D  = request[47]
      next_state.Fastd1W  = request[48]
      next_state.Fastk1W  = request[49]
      next_state.Macd1D   = request[50]
      next_state.Macds1D  = request[51]
      next_state.Macdh1D  = request[52]
      next_state.Macd1W   = request[53]
      next_state.Macds1W  = request[54]
      next_state.Macdh1W  = request[55]
      next_state.Ema81D   = request[56]
      next_state.Ema81W   = request[57]
      next_state.Ema141D  = request[58]
      next_state.Ema141W  = request[59]
      next_state.Ema501D  = request[60]
      next_state.Ema501W  = request[61]
      next_state.Label    = brain.Label

      reward, _ := strconv.ParseFloat(request[62], 64)
      action, _ := strconv.Atoi(request[63])
      brain.Learning(currency, current_state, next_state, reward, action, stateKeyLock)

      conn.Write([]byte(fmt.Sprintf("%d\n", 1)))
    case "BEST_ACTION":
      currency := request[1]

      var current_state state.State
      current_state.Label    = brain.Label
      current_state.Rsi4H    = request[2]
      current_state.EmaRsi4H = request[3]
      current_state.Ema84H   = request[4]
      current_state.Ema144H  = request[5]
      current_state.Ema504H  = request[6]
      current_state.Fastd4H  = request[7]
      current_state.Fastk4H  = request[8]
      current_state.Macd4H   = request[9]
      current_state.Macds4H  = request[10]
      current_state.Macdh4H  = request[11]
      current_state.Rsi1D    = request[12]
      current_state.Rsi1W    = request[13]
      current_state.EmaRsi1D = request[14]
      current_state.EmaRsi1W = request[15]
      current_state.Fastd1D  = request[16]
      current_state.Fastk1D  = request[17]
      current_state.Fastd1W  = request[18]
      current_state.Fastk1W  = request[19]
      current_state.Macd1D   = request[20]
      current_state.Macds1D  = request[21]
      current_state.Macdh1D  = request[22]
      current_state.Macd1W   = request[23]
      current_state.Macds1W  = request[24]
      current_state.Macdh1W  = request[25]
      current_state.Ema81D   = request[26]
      current_state.Ema81W   = request[27]
      current_state.Ema141D  = request[28]
      current_state.Ema141W  = request[29]
      current_state.Ema501D  = request[30]
      current_state.Ema501W  = request[31]

      best_action := brain.ConcurrencyBestAction(currency, current_state, stateKeyLock)
      response := fmt.Sprintf("%f %f\n", best_action[0], best_action[1])

      conn.Write([]byte(response))
    case "SAVE":
      brain.Save(stateKeyLock)
      conn.Write([]byte(fmt.Sprintf("%d\n", 1)))
    case "SETTINGS":
      label     := request[1]
      wallet, _ := strconv.ParseFloat(request[2], 64)
      roi, _    := strconv.ParseFloat(request[3], 64)

      brain.Label = label
      brain.Wallet = wallet
      brain.Roi = roi
      fmt.Println(brain.Label)
      fmt.Println(brain.Wallet)
      fmt.Println(brain.Roi)

      conn.Write([]byte(fmt.Sprintf("%d\n", 1)))
    default:
      conn.Write([]byte("Invalid command. Supported command!\n"))
    }
  }

  if err := scanner.Err(); err != nil {
    log.Printf("Error reading from connection: %v", err)
  }
  
  // fmt.Println("Client disconnected:", conn.RemoteAddr())
}
