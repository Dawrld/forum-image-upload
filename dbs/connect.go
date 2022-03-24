package dbs

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // using driver for sql
)

// conn is holding pool of connetions for database
var conn *sql.DB

// NewConnect ...
func NewConnect() error {
	db, err := sql.Open("sqlite3", "file:forum.s3db?_auth&_auth_user=Dawrld&_auth_pass=Alibi&_auth_crypt=sha256&_foreign_keys=on")
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE  IF NOT EXISTS "user" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"username"	TEXT UNIQUE,
			"email"	TEXT UNIQUE,
			"password"	TEXT
		)`,

		`CREATE TABLE  IF NOT EXISTS "session" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"uid"	INTEGER,
			"uuid"	TEXT UNIQUE,
			"status"	INTEGER DEFAULT 0,
			"datetime"	DATETIME,
			FOREIGN KEY ("uid") REFERENCES user ("id") ON DELETE CASCADE
		)`,

		`CREATE TABLE  IF NOT EXISTS "obj_type" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"name"	TEXT UNIQUE
		)`,

		`CREATE TABLE  IF NOT EXISTS "post" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"uid"	INTEGER,
			"text"	TEXT,
			"image" TEXT,
			"creation_date"	DATETIME,
			FOREIGN KEY ("uid") REFERENCES user ("id") ON DELETE CASCADE
		)`,

		`CREATE TABLE	IF NOT EXISTS "category" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"name"	TEXT UNIQUE
		)`,

		`CREATE TABLE  IF NOT EXISTS "post_category" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"post_id"	INTEGER,
			"category_id"	INTEGER,
			FOREIGN KEY ("post_id") REFERENCES post ("id") ON DELETE CASCADE
			FOREIGN KEY ("category_id") REFERENCES category ("id") ON DELETE CASCADE
			UNIQUE("post_id", "category_id")
		)`,

		`CREATE TABLE  IF NOT EXISTS "comment" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"post_id"	INTEGER,
			"uid"	INTEGER,
			"text"	TEXT,
			"creation_date"	DATETIME,
			FOREIGN KEY ("post_id") REFERENCES post ("id") ON DELETE CASCADE
			FOREIGN KEY ("uid") REFERENCES user ("id") ON DELETE CASCADE
		)`,

		`CREATE TABLE  IF NOT EXISTS "rate_type" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"name"	TEXT UNIQUE
		)`,

		`CREATE TABLE  IF NOT EXISTS "rate" (
			"id"	INTEGER UNIQUE PRIMARY KEY AUTOINCREMENT,
			"rate_type"	INTEGER,
			"obj_type"	INTEGER,
			"uid"	INTEGER,
			"obj_id"	INTEGER,
			FOREIGN KEY ("uid") REFERENCES user ("id") ON DELETE CASCADE
			FOREIGN KEY ("rate_type") REFERENCES rate_type ("id")
			FOREIGN KEY ("obj_type") REFERENCES obj_type ("id")
			UNIQUE ("rate_type","obj_type","uid","obj_id")
		)`,

		`CREATE TRIGGER IF NOT EXISTS "session_update"
			BEFORE INSERT
			ON session
			BEGIN
				UPDATE session SET status = 0 WHERE uid = NEW.uid;
			END;
		`,

		`CREATE TRIGGER IF NOT EXISTS "rate_update"
			BEFORE INSERT ON rate
			WHEN EXISTS (SELECT * FROM rate
				WHERE rate_type != NEW.rate_type
				AND obj_type = NEW.obj_type
				AND uid = NEW.uid
				AND obj_id = NEW.obj_id)
			BEGIN
				UPDATE rate
				SET rate_type = NEW.rate_type
				WHERE obj_type = NEW.obj_type
				AND uid = NEW.uid
				AND obj_id = NEW.obj_id;
				
				SELECT RAISE(IGNORE);
			END;
		`,

		`CREATE TRIGGER IF NOT EXISTS "rate_repeat"
			BEFORE INSERT ON rate
			WHEN EXISTS (SELECT * FROM rate
				WHERE rate_type = NEW.rate_type
				AND obj_type = NEW.obj_type
				AND uid = NEW.uid
				AND obj_id = NEW.obj_id)
			BEGIN
				DELETE FROM rate
				WHERE rate_type = NEW.rate_type
				AND obj_type = NEW.obj_type
				AND uid = NEW.uid
				AND obj_id = NEW.obj_id;

				SELECT RAISE(IGNORE);
			END;
		`,

		`CREATE TRIGGER IF NOT EXISTS "rate_check" BEFORE INSERT ON rate
		WHEN (
			SELECT CASE NEW.obj_type 
				WHEN 1 THEN (SELECT id FROM post WHERE post.id = NEW.obj_id)
				  ELSE (SELECT id FROM comment WHERE comment.id = NEW.obj_id)
			END
		) IS NULL
		BEGIN
		SELECT
		  RAISE(ABORT, 'Object with this ID does not exist');
		END;
		`,
	}

	for _, v := range tables {
		_, err = db.Exec(v)
		if err != nil {
			return err
		}
	}

	fmt.Println("Connected to the database")
	conn = db

	go func() {
		for {
			if err := CleanSessions(); err != nil {
				log.Println(err)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
	return nil
}
