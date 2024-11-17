package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phpgao/tlog"
	"github.com/phpgao/tlog/handler"
	"github.com/robfig/cron/v3"
)

const (
	cronWithDebug = "@every 5s"
	cronRelease   = "@every 1m"
)

type Server struct {
	Router  *gin.Engine
	Config  *Config
	Cron    *cron.Cron
	Dir     string
	MaxSize int
}

func NewServer(conf *Config) *Server {
	return &Server{
		Config: conf,
		Dir:    conf.GetDir(),
		Cron:   cron.New(cron.WithSeconds()),
		Router: gin.New(),
	}
}

func (s *Server) PreCheck() error {
	err := os.MkdirAll(s.Dir, os.ModePerm)
	if err != nil {
		tlog.Errorf("mkdir [%s] error: %s", s.Dir, err)
		return err
	}
	s.MaxSize = s.Config.MaxInt << 20
	tlog.Infof("MaxSize: %dMB", s.MaxSize)
	if !s.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return nil
}

func (s *Server) Run() error {
	err := s.PreCheck()
	if err != nil {
		return err
	}
	s.addCron()
	defer s.stopCron()
	s.addRoute()
	err = s.Router.Run(s.Config.Bind)
	if err != nil {
		tlog.Fatalf("listen [%s] error: %s", s.Config.Bind, err)
	}

	return nil
}

func (s *Server) addRoute() {
	s.Router.Use(handler.GinLogger(), gin.Recovery())
	s.Router.PUT("/:filename", func(c *gin.Context) {
		s.UploadFile(c)
	})
	s.Router.POST("/:filename", func(c *gin.Context) {
		s.UploadFile(c)
	})
	s.Router.PUT("/", func(c *gin.Context) {
		s.UploadFile(c)
	})
	s.Router.POST("/", func(c *gin.Context) {
		s.UploadFile(c)
	})
	s.Router.GET("/:path/:filename", func(c *gin.Context) {
		filePath := c.Param("path")
		fileName := c.Param("filename")
		realPath := filepath.Join(s.Dir, filePath, fileName)
		c.File(realPath)
	})
}

func (s *Server) UploadFile(c *gin.Context) {
	length, err := strconv.Atoi(c.GetHeader("Content-Length"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid content length")
		return
	}
	if length > s.MaxSize {
		c.String(http.StatusRequestEntityTooLarge, "file too large,current %dMB, max %dMB", length>>20, s.MaxSize)
		return
	}
	fileName := c.Param("filename")
	if fileName == "" || len(fileName) > 255 {
		c.String(http.StatusBadRequest, "invalid file name")
		return
	}
	dstPath, downloadPath, err := s.genFilePath(fileName)
	tlog.Infof("dstPath: %s, downloadPath: %s", dstPath, downloadPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "can not create file")
		return
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "can not create file")
		return
	}
	defer func(dst *os.File) {
		_ = dst.Close()
	}(dst)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(s.MaxSize))
	if _, err := io.Copy(dst, c.Request.Body); err != nil {
		c.String(http.StatusInternalServerError, "file copy error")
		tlog.Errorf("copy error: %s", err)
		return
	}
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	wgetUrl := fmt.Sprintf("%s://%s/%s", protocol, c.Request.Host, downloadPath)
	curlUrl := fmt.Sprintf("curl -O %s://%s/%s", protocol, c.Request.Host, downloadPath)
	c.String(http.StatusOK, fmt.Sprintf("\n%s upload \n%s\n%s\n", fileName, wgetUrl, curlUrl))
}

func (s *Server) genFilePath(fileName string) (string, string, error) {
	randomDir := RandStringBytes(s.Config.Length)
	uploadFolder := filepath.Join(s.Dir, randomDir)
	err := os.MkdirAll(uploadFolder, os.ModePerm)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(uploadFolder, fileName), filepath.Join(randomDir, fileName), nil
}

func (s *Server) addCron() {
	var every string
	if s.Config.Debug {
		every = cronWithDebug
	} else {
		every = cronRelease
	}
	s.Cron.AddFunc(every, s.ScanAndDelete)
	s.Cron.Start()
}

func (s *Server) stopCron() {
	s.Cron.Stop()
}

func (s *Server) ScanAndDelete() {
	err := CleanupOldFilesAndEmptyDirs(s.Dir, time.Duration(s.Config.Expire)*24*time.Hour)
	if err != nil {
		tlog.Errorf("scan and delete error: %s", err)
		return
	}
}
