package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"log"
	"context"
	"github.com/joho/godotenv"
)

func InitDB() *pgxpool.Pool {
	err := godotenv.Load("../.env") 
	if err != nil{
		log.Fatal("error to load .env file", err)
	}

	DB, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("failed create pool ", err)
	}

	if err = DB.Ping(context.Background()); err != nil{
		log.Fatal("failed connect db",err)
	}
	
	log.Print("DB connect sucess")
	return DB
}