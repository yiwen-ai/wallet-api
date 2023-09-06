package conf

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"github.com/teambition/gear"
)

// Config ...
var Config ConfigTpl

var AppName = "wallet-api"
var AppVersion = "0.1.0"
var BuildTime = "unknown"
var GitSHA1 = "unknown"

var once sync.Once

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err.Error())
	}

	p := &Config
	readConfig(p, "../../config/default.toml")
	if err := p.Validate(); err != nil {
		panic(err)
	}
	p.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	p.GlobalSignal = gear.ContextWithSignal(context.Background())

	var cancel context.CancelFunc
	p.GlobalShutdown, cancel = context.WithCancel(context.Background())
	go func() {
		<-p.GlobalSignal.Done()
		time.AfterFunc(time.Duration(p.Server.GracefulShutdown)*time.Second, cancel)
	}()
}

type Logger struct {
	Level string `json:"level" toml:"level"`
}

type Server struct {
	Addr             string `json:"addr" toml:"addr"`
	GracefulShutdown uint   `json:"graceful_shutdown" toml:"graceful_shutdown"`
}

type Base struct {
	Userbase   string `json:"userbase" toml:"userbase"`
	Logbase    string `json:"logbase" toml:"logbase"`
	Taskbase   string `json:"taskbase" toml:"taskbase"`
	Walletbase string `json:"walletbase" toml:"walletbase"`
}

type Redis struct {
	Prefix string `json:"prefix" toml:"prefix"`
	Node   string `json:"node" toml:"node"`
}

type Stripe struct {
	PubKey     string `json:"pub_key" toml:"pub_key"`
	PriceID    string `json:"price_id" toml:"price_id"`
	SuccessUrl string `json:"success_url" toml:"success_url"`
	SecretKey  string
	WebhookKey string
}

// ConfigTpl ...
type ConfigTpl struct {
	Rand           *rand.Rand
	GlobalSignal   context.Context
	GlobalShutdown context.Context
	Env            string `json:"env" toml:"env"`
	Logger         Logger `json:"log" toml:"log"`
	Server         Server `json:"server" toml:"server"`
	Redis          Redis  `json:"redis" toml:"redis"`
	Base           Base   `json:"base" toml:"base"`
	Stripe         Stripe `json:"stripe" toml:"stripe"`

	globalJobs int64 // global async jobs counter for graceful shutdown
}

func (c *ConfigTpl) Validate() error {
	c.Stripe.SecretKey = os.Getenv("STRIPE_SECRET_KEY")
	c.Stripe.WebhookKey = os.Getenv("STRIPE_WEBHOOK_SECRET")
	if c.Stripe.SecretKey == "" {
		log.Println("STRIPE_SECRET_KEY is not set")
	}
	return nil
}

func (c *ConfigTpl) ObtainJob() {
	atomic.AddInt64(&c.globalJobs, 1)
}

func (c *ConfigTpl) ReleaseJob() {
	atomic.AddInt64(&c.globalJobs, -1)
}

func (c *ConfigTpl) JobsIdle() bool {
	return atomic.LoadInt64(&c.globalJobs) <= 0
}

func readConfig(v interface{}, path ...string) {
	once.Do(func() {
		filePath, err := getConfigFilePath(path...)
		if err != nil {
			panic(err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		_, err = toml.Decode(string(data), v)
		if err != nil {
			panic(err)
		}
	})
}

func getConfigFilePath(path ...string) (string, error) {
	// 优先使用的环境变量
	filePath := os.Getenv("CONFIG_FILE_PATH")

	// 或使用指定的路径
	if filePath == "" && len(path) > 0 {
		filePath = path[0]
	}

	if filePath == "" {
		return "", fmt.Errorf("config file not specified")
	}

	return filePath, nil
}
