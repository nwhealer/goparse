package main

import (
	"fmt"
	"os"
	"encoding/csv"
	"time"
	"io/ioutil"
	"strings"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

const batchSize int64 = 100
// file headers
// id,first_name,full_name,email,phone_number,address_city,address_street,address_house,address_entrance,address_floor,address_office,address_comment,location_latitude,location_longitude,amount_charged,user_id,user_agent,created_at,address_doorcode

const csvRes string = ".csv"

func main() {
	_t := time.Now() ////

	db := dbConnect()
	flist := getCsvList(".")

	for _, file := range flist {
		go dataInsert(file, db, _t)
	}

	var input string
	fmt.Scanln(&input)
	fmt.Println(time.Now().Sub(_t)) ////
}

func dataInsert(file string, db *sqlx.DB, t time.Time) {
	fileOpen, err := os.Open(file)
	if err != nil {
		fmt.Println("File open error |", file, err)
		return
	}
	defer fileOpen.Close()

	reader := csv.NewReader(fileOpen)

	// first line - headers
	_, err = reader.Read()
	if err != nil {
		fmt.Println("CSV reading error (header) |", file, err)
	}

	corr := []Corruption{}
	var cid int64 = 1

	for {
		line, err := reader.Read()
		if err != nil {
			l := len(line)
			if 0 == l {
				fmt.Printf("File %s parsed and load\n", file)
				break
			}

			fmt.Println("CSV reading error (data string) |", file, l)
			break
		}

		corr = append(corr, MakeCorruption(line))

		cid++

		if cid % batchSize == 0 {
			_, err := db.NamedExec(InsertSql, corr)
			if err != nil {
				fmt.Println("NamedExec error |", err, cid)
				break
			}

			corr = []Corruption{}
		}

	}

	if 0 < len(corr) {
		_, err := db.NamedExec(InsertSql, corr)
		if err != nil {
			fmt.Println("NamedExec error |", err, cid)
		}

		corr = []Corruption{}			
	}

	fmt.Println(time.Now().Sub(t))
}

func dbConnect() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgresql://smf_gps_user:smfgpspassword@10.20.0.4:5432/smf_gps_db?sslmode=disable")
	if err != nil {
		fmt.Println("Connect db error ", err)
	}

	return db
}

// список .CSV фалов в директории
func getCsvList(directory string) map[int]string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	filesList := make(map[int]string)
	var i int = 0

	for _, file := range files {
		if strings.Contains(file.Name(), csvRes) {
			filesList[i] = file.Name()
			i++
		}
	}

	return filesList
}
