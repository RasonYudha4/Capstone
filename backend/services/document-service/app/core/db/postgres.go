package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"log"
	"context"

)

func InitDB()(*pgxpool.Pool, error){
	DB, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("failed create pool ", err)
	}

	if err = DB.Ping(context.Background()); err != nil{
		log.Fatal("failed connect db",err)
	}
	
	log.Print("DB connect sucess")
	return DB, nil
}