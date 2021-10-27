package main

import (
	"bufio"
	"crypto/md5"
	"fmt"

	//"html/template"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	//log
	"go.uber.org/zap"

	//gorm
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	//redis
	"context"
)

const maxSize = 5 << 20 //5MB
// test new ssh key
// type Song struct {
// 	title  string
// 	pic    string
// 	singer string
// 	score  string
// }
type Config struct {
	gorm.Model
	Times int64
}

func getTimes() int64 {
	var times int64 = 1
	//add times
	db, err := gorm.Open("sqlite3", "config.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Config{})
	var sr Config
	db.First(&sr)
	if sr.Times == 0 {
		sr = Config{Times: 1}
		db.Create(&sr)
	} else {
		times = sr.Times + 1
		db.Model(&sr).Update("times", times)
	}

	return times
}

func main() {
	initStart()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	app := iris.New()
	app.RegisterView(iris.HTML("./views", ".html"))
	//app.Use(myMiddleware)
	app.Get("/", func(ctx iris.Context) {
		logger.Info("router info",
			zap.String("method", "GET"),
			zap.String("url", "/"),
		)
		// Bind: {{.message}} with "Hello world!"
		ctx.ViewData("message", "Hi Sayari !")
		// Render template file: ./views/hello.html
		var times int64 = getTimes()
		ctx.ViewData("times", times)
		ctx.View("index.html")
	})

	app.Handle("GET", "/ping", func(ctx iris.Context) {
		logger.Info("router info",
			zap.String("method", "GET"),
			zap.String("url", "/ping"),
		)
		ctx.JSON(iris.Map{"message": "pong"})
	})
	app.HandleDir("/static", "./views")
	// Serve the upload_form.html to the client.
	app.Get("/upload", func(ctx iris.Context) {
		// create a token (optionally).
		logger.Info("router info",
			zap.String("method", "GET"),
			zap.String("url", "/upload"),
		)
		now := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(now, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		// render the form with the token for any use you'd like.
		// ctx.ViewData("", token)
		// or add second argument to the `View` method.
		// Token will be passed as {{.}} in the template.
		ctx.View("upload_form.html", token)
	})
	app.Post("/upload", iris.LimitRequestBodySize(maxSize), func(ctx iris.Context) {
		//fmt.Printf("upload")
		logger.Info("router info",
			zap.String("method", "POST"),
			zap.String("url", "/upload"),
		)
		// Get the file from the request.
		file, info, err := ctx.FormFile("uploadfile")
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.HTML("Error while uploading: <b>" + err.Error() + "</b>")
			return
		}

		defer file.Close()
		fname := info.Filename

		// Create a file with the same name
		// assuming that you have a folder named 'uploads'
		out, err := os.OpenFile("./views/upload/"+fname,
			os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.HTML("Error while uploading: <b>" + err.Error() + "</b>")
			return
		}
		defer out.Close()

		io.Copy(out, file)
		path := "https://loveloliii.monster/static/upload/" + fname
		ctx.Writef("file path is -->:\n%s", path)
	})
	// 使用nginx监听80/443 再转发到8888
	app.Listen(":8888")
}
func beforeSave(ctx iris.Context, file *multipart.FileHeader) {
	// ip := ctx.RemoteAddr()
	// ip = strings.Replace(ip, ".", "_", -1)
	// ip = strings.Replace(ip, ":", "_", -1)
	//file.Filename = ip + "-" + file.Filename
	//fmt.Printf("fileNmae:%s", file.Filename)
}
func myMiddleware(ctx iris.Context) {
	ctx.Application().Logger().Infof("Runs before %s", ctx.Path())
	ctx.Next()
}

func readJSON(filePath string) (result string) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		//		fmt.Println("ERROR:", err)
	}
	buf := bufio.NewReader(file)
	for {
		s, err := buf.ReadString('\n')
		result += s
		if err != nil {
			if err == io.EOF {
				fmt.Println("Read is ok")
				break
			} else {
				fmt.Println("ERROR:", err)
				return
			}
		}
	}
	return result
}

// ScoreMap is a scoreMap.
type ScoreMap struct {
	Key   string
	Value string
}

// Product is a simple
type Product struct {
	gorm.Model
	Code  string
	Price uint
}

// Song is a song.
type Song struct {
	Title  string
	Pic    string
	Singer string
	Score  string
	Issue  string
	gorm.Model
}

func getSong(t string) Song {
	db, err := gorm.Open("sqlite3", "song.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Song{})
	var ss Song
	//db.First(&product, 1)
	db.First(&ss, "title = ?", t)
	return ss
}
func querySong(t string) []Song {
	db, err := gorm.Open("sqlite3", "song.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Song{})
	var ss []Song
	db.Where("title LIKE ?", "%"+t+"%").Find(&ss)
	return ss
}
func addSong(s Song) {
	db, err := gorm.Open("sqlite3", "song.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Song{})
	var sr Song
	db.Where("issue =?", s.Issue).First(&sr)
	fmt.Printf(sr.Issue)
	if sr.Issue == "" {
		db.Create(&Song{Title: s.Title, Pic: s.Pic,
			Singer: s.Singer, Score: s.Score, Issue: s.Issue})
		fmt.Printf("add success")
	} else {
		fmt.Printf("add failed")
	}
}

// do init

func initStart() {
}

var ctxx = context.Background()
