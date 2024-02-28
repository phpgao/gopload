package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/phpgao/gopload/internal/config"
	"github.com/phpgao/gopload/internal/util"
)

const (
	cronWithDebug = "@every 5s"
	cronRelease   = "@every 1m"
)

type Server struct {
	Router  *gin.Engine
	Config  *config.Config
	Cron    *cron.Cron
	Dir     string
	MaxSize int
}

func NewServer(conf *config.Config) *Server {
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
		log.Errorf("mkdir [%s] error: %s", s.Dir, err)
		return err
	}

	s.MaxSize = s.Config.MaxInt << 20
	log.Infof("max size: %d MB", s.MaxSize)
	log.Infof("s.Config.MaxInt size: %d MB", s.Config.MaxInt)
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
		log.WithError(err).Error("error running server")
	}

	return nil
}

func (s *Server) addRoute() {
	s.Router.Use(ginlogrus.Logger(log.New()), gin.Recovery())
	s.Router.PUT("/:filename", func(c *gin.Context) {
		s.UploadFile(c)
	})
	s.Router.POST("/:filename", func(c *gin.Context) {
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
		c.String(http.StatusBadRequest, "无法解析 Content-Length 头部。")
		return
	}
	log.Infof("length: %d", length)
	log.Infof("MaxSize: %d", s.MaxSize)
	if length > s.MaxSize {
		c.String(http.StatusRequestEntityTooLarge, "文件过大。")
		return
	}
	fileName := c.Param("filename")
	dstPath, downloadPath, err := s.genFilePath(fileName)
	log.Infof("dstPath: %s, downloadPath: %s", dstPath, downloadPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "无法创建文件。")
		return
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "文件上传失败。")
		return
	}
	defer dst.Close()
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(s.MaxSize))
	log.Debug(dstPath)
	if _, err := io.Copy(dst, c.Request.Body); err != nil {
		c.String(http.StatusInternalServerError, "文件上传失败。")
		log.Errorf("copy error: %s", err)
		return
	}
	wgetUrl := fmt.Sprintf("wget https://%s/%s", c.Request.Host, downloadPath)
	curlUrl := fmt.Sprintf("curl -O https://%s/%s", c.Request.Host, downloadPath)
	c.String(http.StatusOK, fmt.Sprintf("\n%s 上传成功！\n%s\n%s\n", fileName, wgetUrl, curlUrl))
}

func (s *Server) genFilePath(fileName string) (string, string, error) {
	randomDir := util.RandStringBytes(s.Config.Length)
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
	err := util.CleanupOldFilesAndEmptyDirs(s.Dir, time.Duration(s.Config.Expire)*24*time.Hour)
	if err != nil {
		log.Errorf("scan and delete error: %s", err)
		return
	}
}
