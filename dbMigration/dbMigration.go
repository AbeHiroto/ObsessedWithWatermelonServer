package main

import (
	"fmt"
	"os"

	//"encoding/json"
	//"io/ioutil"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User モデルの定義（修正済み）
type User struct {
	gorm.Model
	HasRoom    bool `gorm:"not null;default:false"`
	HasRequest bool `gorm:"not null;default:false"`
}

// GameRoom モデルの定義
type GameRoom struct {
	gorm.Model
	UserID           uint   `gorm:"not null"`
	RoomCreator      string `gorm:"not null"` // 作成者ニックネーム
	GameState        string `gorm:"not null;default:'created'"`
	UniqueToken      string `gorm:"unique;not null"` // 招待URL
	FinishTime       int64
	StartTime        int64
	RoomTheme        string
	ChallengersCount int          `gorm:"default:0"`             // 申請者数
	Challengers      []Challenger `gorm:"foreignKey:GameRoomID"` // 結びつく入室申請を取得
}

// 挑戦者は別テーブルで管理（複数の挑戦者に対応）
type Challenger struct {
	gorm.Model
	UserID             uint
	GameRoomID         uint     `gorm:"index"` // GameRoomテーブルのIDを参照
	ChallengerNickname string   // 挑戦希望者のニックネーム
	Status             string   `gorm:"index;default:'pending'"` // 申請状態を表す
	GameRoom           GameRoom `gorm:"foreignKey:GameRoomID"`   // GameRoomへの参照
}

var logger *zap.Logger

func init() {
	_ = godotenv.Load()
	// Zapのロガー設定
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// マイグレーションを実行する関数
func AutoMigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(&User{}, &GameRoom{}, &Challenger{})
	if err != nil {
		logger.Error("Error migrating tables", zap.Error(err))
	} else {
		logger.Info("User and GameRoom tables created successfully")
	}

	// 複合インデックスの作成
	err = db.Exec("CREATE INDEX IF NOT EXISTS idx_user_id_game_state ON game_rooms (user_id, game_state)").Error
	if err != nil {
		logger.Error("Error creating index idx_user_id_game_state", zap.Error(err))
	}
}

func main() {
	// 環境変数からデータベースの接続情報を取得
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s", host, user, dbname, password, sslmode)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return
	}

	// データベース接続の取得
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Error("Failed to get SQLDB", zap.Error(err))
		return
	}
	defer sqlDB.Close()

	// マイグレーション実行
	AutoMigrateDB(gormDB)
}
