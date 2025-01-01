package state

import "sync"

type State struct {
  Rsi4H    string
  EmaRsi4H string
  Ema84H   string
  Ema144H  string
  Ema504H  string
  Fastd4H  string
  Fastk4H  string
  Macd4H   string
  Macds4H  string
  Macdh4H  string
  Rsi1D    string
  Rsi1W    string
  EmaRsi1D string
  EmaRsi1W string
  Fastd1D  string
  Fastk1D  string
  Fastd1W  string
  Fastk1W  string
  Macd1D   string
  Macds1D  string
  Macdh1D  string
  Macd1W   string
  Macds1W  string
  Macdh1W  string
  Ema81D   string
  Ema81W   string
  Ema141D  string
  Ema141W  string
  Ema501D  string
  Ema501W  string
  Label    string
}

type StateKeyLock struct {
  locks map[string]*sync.RWMutex

  mapLock sync.RWMutex
}

func NewStateKeyLock() *StateKeyLock {
  return &StateKeyLock{locks: make(map[string]*sync.RWMutex)}
}

// writer
func (l *StateKeyLock) getLockBy(key string) *sync.RWMutex {
  l.mapLock.Lock()                   
  defer l.mapLock.Unlock()           
                                     
  ret, found := l.locks[key]
  if found {
    return ret
  }

  ret = &sync.RWMutex{}
  l.locks[key] = ret
  return ret
}

func (l *StateKeyLock) Lock(key string) {
  l.getLockBy(key).Lock()
}

func (l *StateKeyLock) Unlock(key string) {
  l.getLockBy(key).Unlock()
}

// reader
func (l *StateKeyLock) getRLockBy(key string) *sync.RWMutex {
  l.mapLock.RLock()
  defer l.mapLock.RUnlock()           
                                     
  ret, found := l.locks[key]
  if found {
    return ret
  }

  ret = &sync.RWMutex{}
  l.locks[key] = ret
  return ret
}

func (l *StateKeyLock) RLock(key string) {
  l.getLockBy(key).RLock()
}

func (l *StateKeyLock) RUnlock(key string) {        
  l.getLockBy(key).RUnlock()                       
}

func (state *State) LockKey(currency string) string {
  return "key" // 13, 14, 14
  // return state.CurrencyId(currency) + "key" // 13, 14, 14
  // return state.CurrencyId(currency) + state.WorkerId + "key" // 13, 14, 14
  // return state.CurrencyId(currency) + "key" // 13, 13, 
  // return state.CurrencyId(currency) + state.Mfi + state.Fastd + state.Fastk
  // return state.CurrencyId(currency) + state.Mfi[:1] + state.Fastd[:1] + state.Fastk[:1]
}

func (state *State) CurrencyId(currency string) string {
  switch currency {
    case "sol":
      return "1"
    case "ada":
      return "2"
    case "matic":
      return "3"
    case "gala":
      return "4"
    case "mana":
      return "5"
    default:
      return "0"
    }
}