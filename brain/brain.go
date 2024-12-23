package brain

import (
	"context"
	"fmt"
	"q-learning/state"
	"reflect"
	"strconv"

	"github.com/redis/go-redis/v9"
	// "sync"
	// "strconv"
	// "sort"
)

type Brain struct {
  QTable map[state.State][2]float64
  Label string // 1 - alta forte
  Wallet float64
  Roi float64
}

func (brain *Brain) ConcurrencyBestAction(currency string, state state.State, stateKeyLock *state.StateKeyLock) [2]float64 {
  lockKey := state.LockKey(currency)
  defer stateKeyLock.RUnlock(lockKey)

  stateKeyLock.RLock(lockKey)
  actions, qExist := brain.QTable[state]
  if (qExist == false) {
    return [2]float64{0.0, 0.0}
  }

  if(actions[0] >= actions[1]) {
    return [2]float64{0.0, actions[0]}
  } else {
    return [2]float64{1.0, actions[1]}
  }
}

func (brain *Brain) BestAction(state state.State) [2]float64 {
  actions, qExist := brain.QTable[state]

  if (qExist == false) {
    return [2]float64{0.0, 0.0}
  }

  if(actions[0] >= actions[1]) {
    return [2]float64{0.0, actions[0]}
  } else {
    return [2]float64{1.0, actions[1]}
  }
}

func (brain *Brain) Learning(currency string, state state.State, nextState state.State, reward float64, action int, stateKeyLock *state.StateKeyLock) {
  lockKey := state.LockKey(currency)

  defer stateKeyLock.Unlock(lockKey)

  stateKeyLock.Lock(lockKey)
  actions, qExist := brain.QTable[state]

  best_action := brain.BestAction(nextState)

  if (qExist == false) {
    actions = [2]float64{0.0, 0.0}
  }

  // Learning Rate (LearningRate)
  //  Definition: Determines how much the new information overrides the old information in the Q-value update.
  //  Range: Typically between 0.0 and 1.0.
  // Scenario                     Value Range Explanation
  //  Slow learning, stable env.  0.01 - 0.1  Small updates prevent overreaction to noisy rewards.
  //  Balanced learning           0.1  - 0.3  Common for most scenarios, balances stability and speed.
  //  Fast learning, dynamic env. 0.5  - 1.0  Quickly adapts to changes but risks instability.
  // -------------------------
  // Discount Rate (DiscountRate)
  //  Definition: Determines the importance of future rewards compared to immediate rewards.
  //  Range: Typically between 0.0 and 1.0.
  // Scenario                   Value Range Explanation
  //  Short-term focus          0.1 - 0.3   Immediate rewards dominate; used in fast-paced tasks.
  //  Balanced long-term focus  0.5 - 0.8   Mix of short-term and future rewards; common default.
  //  Long-term focus           0.9 - 1.0   Prioritizes future rewards; ideal for long-term planning.

  gamma         := 0.50 // 0.90 // 0.95 -> Valores utilizados no chatGpt *DiscountRate
  learning_rate := 0.20 // 0.08 // 0.10 -> Valores utilizados no chatGpt *LearningRate

  q_target := reward + gamma * best_action[1]
  q_error := q_target - actions[action]
  actions[action] = actions[action] + (learning_rate * q_error)

  // newQ := currentQ + LearningRate*(reward+DiscountRate*maxNextQ-currentQ)
  brain.QTable[state] = actions
}

func (brain *Brain) Save(stateKeyLock *state.StateKeyLock) {
  stateKeyLock.Lock("save")

  // ruby
  /*redis = Redis.new(url: "redis://192.168.0.8:6379", db: "3")
  keys = redis.keys("brain:*")
  resp = keys.map do |key|
    _, id = key.split(":")
    if id[0] == "2"
      "#{id} -> #{redis.get(key)}"
    else
      ""
    end
  end
  puts resp.uniq.join("\n")*/

  client := redis.NewClient(&redis.Options{
    Addr:   "freqai.pro:6379",
    Password: "", // no password set
    DB:     3,  // use default DB
  })

  ctx := context.Background()
  for state, value := range brain.QTable {
    // adicionar currency e label
    var redis_key = ""

    values := reflect.ValueOf(state)
    for i := 0; i < values.NumField(); i++ {
      if(i == 0) {
        redis_key = "brain:" + fmt.Sprintf("%v",values.Field(i))
      } else {
        redis_key = redis_key + "|" + fmt.Sprintf("%v",values.Field(i))
      }
    }

    value_key := strconv.FormatFloat(value[0], 'f', -1, 64)
    value_key = value_key + "|" + strconv.FormatFloat(value[1], 'f', -1, 64)

    client.Set(ctx, redis_key, value_key, 0).Err()
    // fmt.Printf("key[%s] value[%s]\n", redis_key, value_key)
  }

  fmt.Printf("------ len(brain.QTable) -> %d------\n", len(brain.QTable))
  stateKeyLock.Unlock("save")
}
