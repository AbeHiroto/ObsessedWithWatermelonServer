package main

import (
	"net/http"
	//"os"
	"time"

	"go.uber.org/zap"

	"xicserver/database" //PostgreSQLとRedisの初期化
	"xicserver/handlers" //Websocket接続へのアップグレードとホーム画面での構成に必要な情報の取得
	"xicserver/models"   //モデル定義
	"xicserver/screens"  //フロントの画面構成やマッチングに関連するHTTPリクエストの処理
	"xicserver/utils"    //ロガーの初期化とCronジョブ(PostgreSQLの定期クリーンナップ)

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	//"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var logger *zap.Logger
	var err error

	// ロガーの初期化とクリーンナップ
	logger, err = utils.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// .envファイルを読み込む
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	// Websocket接続で用いる変数を初期化
	clients := make(map[*models.Client]bool)
	games := make(map[uint]*models.Game)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 非同期でPostgreSQLとRedisの初期化
	var db *gorm.DB
	var rdb *redis.Client
	done := make(chan bool)

	go func() {
		// データベース接続情報を環境変数から取得
		db, err = database.InitPostgreSQL(logger)
		if err != nil {
			logger.Fatal("PostgreSQLの初期化に失敗しました", zap.Error(err))
		}
		done <- true
	}()

	go func() {
		rdb, err = database.InitRedis(logger)
		if err != nil {
			logger.Fatal("Failed to initialize Redis", zap.Error(err))
		}
		done <- true
	}()

	// 2つの初期化が完了するのを待つ
	<-done
	<-done

	// クーロンスケジューラのセットアップと呼び出し
	go utils.CronCleaner(db, logger)

	router := gin.Default()
	// dbとrdbを全てのリクエストで利用できるようにする
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("rdb", rdb)
		c.Next()
	})
	//リクエストロガーを起動
	router.Use(gin.Recovery(), utils.RequestLogger(logger))

	//CORS（Cross-Origin Resource Sharing）ポリシーを設定
	router.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"https://abehiroto.com"}, // 本番環境用設定
		// ローカルテスト時は"AllowOrigins: []string{"https://abehiroto.com", "http://localhost:42951", "http://localhost:35441"},"のようにシミュレーター起動後にlocalhostのポート指定
		AllowOrigins:     []string{"*"}, // ですべての接続元を許可（ローカル環境のみで適用すること）
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// OPTIONSリクエストを処理するハンドラ
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	//各HTTPリクエストのルーティング
	router.POST("/create", func(c *gin.Context) {
		screens.NewGame(c, db, logger)
	})
	router.GET("/home", func(c *gin.Context) {
		handlers.HomeHandler(c, db, logger)
	})
	router.GET("/room/info", func(c *gin.Context) {
		screens.MyRoomInfo(c, db, logger)
	})
	router.PUT("/request/reply", func(c *gin.Context) {
		screens.ReplyHandler(c, db, logger)
	})
	router.DELETE("/room", func(c *gin.Context) {
		screens.DeleteMyRoom(c, db, logger)
	})
	router.GET("/play/:uniqueToken", func(c *gin.Context) {
		screens.GetRoomInfo(c, db, logger)
	})
	router.POST("/challenger/create/:uniqueToken", func(c *gin.Context) {
		screens.NewChallenge(c, db, logger)
	})
	router.GET("/request/info", func(c *gin.Context) {
		screens.MyRequestInfo(c, db, logger)
	})
	router.DELETE("/request/disable", func(c *gin.Context) {
		screens.DisableMyRequest(c, db, logger)
	})
	// // 本番環境ではコメントアウト解除
	// router.GET("/wss", func(c *gin.Context) {
	// 	handlers.WebSocketConnections(c.Request.Context(), c.Writer, c.Request, db, rdb, logger, clients, games, upgrader)
	// })
	router.GET("/ws", func(c *gin.Context) {
		handlers.WebSocketConnections(c.Request.Context(), c.Writer, c.Request, db, rdb, logger, clients, games, upgrader)
	})

	// テスト時はHTTPサーバーとして運用。デフォルトポートは ":8080"
	router.Run()

	// // 本番環境ではコメントアウトを解除し、HTTPSサーバーとして運用
	// err = router.RunTLS(":443", "/pathto/cert.pem", "/pathto/privkey.pem")
	// if err != nil {
	// 	logger.Fatal("Failed to run HTTPS server: ", zap.Error(err))
	// }
}
