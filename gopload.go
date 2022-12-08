package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/urfave/cli/v2"
)

// random letter const
const (
	letterBytes   = "_0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())
var conf config
var log = logrus.New()

// var debug = true

type config struct {
	Bind    string
	Debug   bool
	MaxInt  int
	MaxSize int
	Length  int
	Dir     string
	Path    string
	Http    bool
}

func init() {
	conf = config{}
}

func preCheck() {
	if conf.Dir == "" {
		conf.Path = filepath.Join(os.TempDir(), "gopload")
	} else {
		var err error
		conf.Path, err = filepath.Abs(conf.Dir)
		if err != nil {
			panic(err)
		}
	}
	log.Infof("gopload path: %s", conf.Path)
	err := os.MkdirAll(conf.Path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	conf.MaxSize = conf.MaxInt << 20

	if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	go addCron()
}

func main() {
	app := &cli.App{
		Name:   "gopload",
		Usage:  "upload file via cli",
		Action: gopload,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "listen",
				Usage:       "0.0.0.0:8088 or :8088",
				Value:       ":8088",
				Destination: &conf.Bind,
				EnvVars:     []string{"SERVER_ADDR"},
				Aliases:     []string{"b"},
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "dir for storing file",
				Value:       "",
				EnvVars:     []string{"SERVER_DIR"},
				Destination: &conf.Dir,
			},
			&cli.BoolFlag{
				Name:        "debug mode",
				Usage:       "debug",
				EnvVars:     []string{"DEBUG"},
				Destination: &conf.Debug,
			},
			&cli.IntFlag{
				Name:        "dir-length",
				Value:       7,
				Destination: &conf.Length,
				EnvVars:     []string{"DRI_LENGTH"},
				Aliases:     []string{"l"},
			},
			&cli.IntFlag{
				Name:        "max",
				Usage:       "max file size in MB",
				Value:       100,
				EnvVars:     []string{"MAX_SIZE"},
				Destination: &conf.MaxInt,
				Aliases:     []string{"m"},
			},
			&cli.BoolFlag{
				Name:        "k",
				Usage:       "disable https",
				EnvVars:     []string{"USE_HTTP"},
				Destination: &conf.Http,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func gopload(c *cli.Context) error {
	preCheck()
	router := addRoute()
	log.Infof("gopload bind: %s", conf.Bind)
	err := router.Run(conf.Bind)
	if err != nil {
		log.WithError(err).Error("error running server")
	}
	return nil
}

func addRoute() *gin.Engine {
	router := gin.New()
	router.Use(ginlogrus.Logger(log), gin.Recovery())

	router.PUT("/:filename", func(c *gin.Context) {
		length, _ := strconv.Atoi(c.GetHeader("Content-Length"))
		if length > conf.MaxSize {
			c.String(http.StatusForbidden, "file too large")
			return
		}
		//domain
		domain := c.Request.Host
		var schema string
		if conf.Http {
			schema = "http"
		} else {
			schema = "https"
		}
		fileName := c.Param("filename")
		dstPath, downloadPath, err := genFilePath(fileName)
		if err != nil {
			panic(err)
		}
		dst, err := os.Create(dstPath)
		if err != nil {
			panic(err)
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(conf.MaxSize))
		log.Debug(dstPath)
		_, _ = io.Copy(dst, c.Request.Body)
		wget := fmt.Sprintf("wget %s://%s/%s", schema, domain, downloadPath)
		curl := fmt.Sprintf("curl -O %s://%s/%s", schema, domain, downloadPath)
		c.String(http.StatusOK, fmt.Sprintf("\n%s uploaded!\n%s\n%s\n", fileName, wget, curl))
	})
	router.GET("/:path/:filename", func(c *gin.Context) {
		filePath := c.Param("path")
		fileName := c.Param("filename")
		realPath := filepath.Join(conf.Path, filePath, fileName)
		c.File(realPath)
	})
	return router
}

// genFilePath return a absolute path with given filename
func genFilePath(fileName string) (string, string, error) {
	var i = 0
	for {
		if i > 100 {
			panic("dir dry")
		}
		randomDir := randStringBytesMaskImprSrcUnsafe(conf.Length)
		uploadFolder := filepath.Join(conf.Path, randomDir)
		if _, err := os.Stat(filepath.Join(uploadFolder, fileName)); err == nil {
			log.Debug("continue")
			i++
			continue
		}
		err := os.MkdirAll(uploadFolder, os.ModePerm)
		if err != nil {
			return "", "", err
		}
		return filepath.Join(uploadFolder, fileName), filepath.Join(randomDir, fileName), nil
	}

}

// randStringBytesMaskImprSrcUnsafe generate random dir with given string length
func randStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

// addCron ...
func addCron() {
	c := cron.New(cron.WithSeconds())
	if conf.Debug {
		c.AddFunc("@every 5s", scanAndDelete)
	} else {
		c.AddFunc("@every 1m", scanAndDelete)
	}
	c.Start()
}

// scanAndDelete run cron to delete outdate file and dir
func scanAndDelete() {
	log.Debug("run cron ..")
	err := filepath.Walk(conf.Path, deleteFile)
	if err != nil {
		log.Debug("delete fail")
		return
	}
}

// deleteFile delete outdated file
func deleteFile(path string, f os.FileInfo, err error) (e error) {
	if f.IsDir() {
		processDir(path, f)
	} else {
		processFile(path, f)
	}

	return
}

// processFile delete outdated file
func processFile(path string, info os.FileInfo) {
	diff := time.Since(info.ModTime())

	if conf.Debug {
		log.Infof("file %s removed\n", path)
		os.Remove(path)
		return
	}
	// 3 days ago
	if diff > time.Second*60*60*24*3 {
		log.Infof("file %s removed\n", path)
		os.Remove(path)
	}
}

// delete empty dir
func processDir(path string, info os.FileInfo) {
	if path == conf.Path {
		return
	}

	if isDirEmpty(path) {
		log.Infof("dir %s removed\n", path)
		err := os.Remove(path)
		if err != nil {
			panic(err)
		}
	}
}

func isDirEmpty(name string) bool {
	f, _ := os.Open(name)
	defer f.Close()
	_, err := f.Readdir(1)
	return err == io.EOF
}
