package main

import (
	"fmt"
	"os"
	"encoding/csv"
	"time"
	"strconv"

	"flist"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

type Corruption struct {
	Id int64 `db:"id"`
	YaId int64 `db:"ya_id"`
	FirstName string `db:"first_name"`
	FullName string `db:"full_name"`
	Email string `db:"email"`
	PhoneNumber int64 `db:"phone_number"`
	City string `db:"address_city"`
	Street string `db:"address_street"`
	House string `db:"address_house"`
	Entrance string `db:"address_entrance"`
	Floor string `db:"address_floor"`
	Office string `db:"address_office"`
	AddressComment string `db:"address_comment"`
	Doorcode string `db:"address_doorcode"`
	Location string `db:"location"`
	Latitude string `db:"location_latitude"`
	Longitude string `db:"location_longitude"`
	AmountCharged string `db:"amount_charged"`
	UserId int64 `db:"user_id"`
	UserAgent string `db:"user_agent"`
	CreatedAt string `db:"created_at"`
}

const batchSize int64 = 100
const comma rune = ','

func main() {
	_t := time.Now() ////

	db := dbConnect()
	flist := flist.GetList(".")

	var cid int64
	err := db.Get(&cid, "SELECT MAX(id)+1 FROM corruption")
	if err != nil {
		cid = 1
		fmt.Println("Get max Id error |", err)
	}

	// не ест этот формат запроса
	// NamedExec(sql, slice) выдаёт pq: got XXXX parameters but the statement requires YY
	// 	insertSql := `INSERT INTO corruption VALUES(
	// :id,
	// :ya_id,
	// :first_name,
	// :full_name,
	// :email,
	// :phone_number,
	// :address_city,
	// :address_street,
	// :address_house,
	// :address_entrance,
	// :address_floor,
	// :address_office,
	// :address_comment,
	// :address_doorcode,
	// :location,
	// :location_latitude,
	// :location_longitude,
	// :amount_charged,
	// :user_id,
	// :user_agent,
	// :created_at
	// )`
	corr := []Corruption{}

	insertSql := `INSERT INTO corruption
(id,
ya_id,
first_name,
full_name,
email,
phone_number,
address_city,
address_street,
address_house,
address_entrance,
address_floor,
address_office,
address_comment,
address_doorcode,
location,
location_latitude,
location_longitude,
amount_charged,
user_id,
user_agent,
created_at) VALUES(
:id,
:ya_id,
:first_name,
:full_name,
:email,
:phone_number,
:address_city,
:address_street,
:address_house,
:address_entrance,
:address_floor,
:address_office,
:address_comment,
:address_doorcode,
:location,
:location_latitude,
:location_longitude,
:amount_charged,
:user_id,
:user_agent,
:created_at
)`

	for _, file := range flist {
		fileOpen, err := os.Open(file)
		if err != nil {
			fmt.Println("File open error |", file, err)
			continue
		}
		defer fileOpen.Close()

		reader := csv.NewReader(fileOpen)
		//reader.Comma = comma

		// first line - headers
		_, err = reader.Read()
		if err != nil {
			fmt.Println("CSV reading error (header) |", file, err)
		}

		for {
			line, err := reader.Read()
			if err != nil {
				fmt.Println("CSV reading error (data string) |", file, err)
				break
			}

			corr = append(corr, Corruption{
				cid,
				strToInt(line[0]),
				line[1],
				line[2],
				line[3],
				strToInt(line[4]),
				line[5],
				line[6],
				line[7],
				line[8],
				line[9],
				line[10],
				line[11],
				line[18],
				fmt.Sprintf("(%s, %s)", line[12], line[13]),
				line[12],
				line[13],
				line[14],
				strToInt(line[15]),
				line[16],
				line[17]})

			cid++

			if cid % batchSize == 0 {
				_, err := db.NamedExec(insertSql, corr)
				if err != nil {
					fmt.Println("NamedExec error |", err, cid)
					break
				}

				corr = []Corruption{}
			}

		}

		if 0 < len(corr) {
			_, err := db.NamedExec(insertSql, corr)
			if err != nil {
				fmt.Println("NamedExec error |", err, cid)
			}

			corr = []Corruption{}			
		}
	}

	fmt.Println(time.Now().Sub(_t)) ////
}

func strToInt(str string) int64 {
	intFromStr, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Convert values to int error |", err)
		return 0
	}

	return int64(intFromStr)
}

func dbConnect() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgresql://smf_gps_user:smfgpspassword@10.20.0.4:5432/smf_gps_db?sslmode=disable")
	if err != nil {
		fmt.Println("Connect db error ", err)
	}

	return db
}