package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQL Table Creation Constants

const createShowtimesTable = `
    CREATE TABLE IF NOT EXISTS showtimes (
        showtime_id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        date DATE NOT NULL,
        showtime TIME NOT NULL,
        theater TEXT NOT NULL,
        price NUMERIC(4,2) NOT NULL,
        admin_id INTEGER NOT NULL,
        FOREIGN KEY (admin_id) REFERENCES admins (admin_id),
        CHECK (date LIKE '____-__-__'),
        CHECK (showtime LIKE '__:__ AM' OR showtime LIKE '__:__ PM'),
        CONSTRAINT uq_showtime UNIQUE (date, showtime, theater)
    )
`

const createTicketsTable = `
    CREATE TABLE IF NOT EXISTS tickets (
        ticket_id INTEGER PRIMARY KEY AUTOINCREMENT,
        showtime_id INTEGER NOT NULL,
        date DATE NOT NULL,
        showtime TIME NOT NULL,
        title TEXT NOT NULL,
        available TEXT NOT NULL DEFAULT 'yes',
        theater TEXT NOT NULL,
        row CHAR(1) NOT NULL,
        seat INTEGER NOT NULL,
        price NUMERIC(4,2) NOT NULL,
        FOREIGN KEY (showtime_id) REFERENCES showtimes (showtime_id)
    )
`

const createAdminsTable = `
    CREATE TABLE IF NOT EXISTS admins (
        admin_id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`

// Create a sales table

func admin_login(db *sql.DB) int {
	var username, password string
	fmt.Println("\n...Admin Login Credentials...")
	fmt.Print("Enter username: ")
	fmt.Scanln(&username)

	fmt.Print("Enter password: ")
	fmt.Scanln(&password)

	row := db.QueryRow("SELECT admin_id FROM admins WHERE username=? AND password=?", username, password)
	var adminID int
	err := row.Scan(&adminID)
	if err != nil {
		log.Fatal(err)
	}
	if adminID != 0 {
		fmt.Println("Admin Login Successful")
		return adminID
	} else {
		fmt.Println("Incorrect username or password")
		return 0
	}
}

func admin_abilities(db *sql.DB, adminID int) {
	for {
		// Prompt admin for input
		var option int
		fmt.Println("\nPlease select an option:")
		fmt.Println("1. Add Movie")
		fmt.Println("2. Quit Admin Duties")
		fmt.Print("Option: ")
		fmt.Scanln(&option)

		switch option {
		case 1:
			add_movie(db, adminID)
		case 2:
			fmt.Println("Exiting Admin Duties...")
			return // Exit the admin_abilities function
		default:
			fmt.Println("Invalid option selected")
		}
	}
}

func add_movie(db *sql.DB, adminID int) {
	var title, date, showtime, theaterStr string
	var price float64
	var n_showings int

	fmt.Println("\nLet's Add a Movie")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Title: ")
	title, _ = reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("How Many Showings: ")
	nShowingsStr, _ := reader.ReadString('\n')
	nShowingsStr = strings.TrimSpace(nShowingsStr)
	n_showings, _ = strconv.Atoi(nShowingsStr)

	for i := 1; i <= n_showings; i++ {
		fmt.Printf("\nShowtime Info for Showing %d\n", i)

		// Get valid date input from user
		for {
			fmt.Print("Date (yyyy-mm-dd): ")
			date, _ = reader.ReadString('\n')
			date = strings.TrimSpace(date)

			_, err := time.Parse("2006-01-02", date)
			if err == nil {
				break
			} else {
				fmt.Println("Invalid date format. Please enter a date in the format yyyy-mm-dd.")
			}
		}

		// Get valid showtime input from user
		for {
			fmt.Print("Showtime (hh:mm AM/PM): ")
			showtime, _ = reader.ReadString('\n')
			showtime = strings.TrimSpace(showtime)

			_, err := time.Parse("03:04 PM", showtime)
			if err == nil {
				break
			} else {
				fmt.Println("Invalid showtime format. Please enter a showtime in the format hh:mm AM/PM.")
			}
		}

		// Get valid theater input from user
		for {
			fmt.Print("Theater (1-6): ")
			theaterStr, _ = reader.ReadString('\n')
			theaterStr = strings.TrimSpace(theaterStr)

			theater, err := strconv.Atoi(theaterStr)
			if err != nil || theater < 1 || theater > 6 {
				fmt.Println("Invalid theater number. Please enter a number between 1 and 6.")
			} else {
				break
			}
		}

		fmt.Print("Ticket Price: ")
		priceStr, _ := reader.ReadString('\n')
		priceStr = strings.TrimSpace(priceStr)

		var err error
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Println("Invalid price entered. Please enter a valid decimal number.")
			continue
		}

		// Use prepared statement to insert showtime into database
		// Assure that this is valid?????
		stmt, err := db.Prepare(`INSERT INTO showtimes (title, date, showtime, theater, price, admin_id)
                            	VALUES (?, ?, ?, ?, ?, ?)
								ON CONFLICT(date, showtime, theater) DO NOTHING`)
		if err != nil {
			fmt.Println("Error preparing statement:", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(title, date, showtime, theaterStr, price, adminID)
		if err != nil {
			fmt.Println("Error inserting showtime:", err)
			return
		}

		// Get the ID of the inserted showtime
		var showtime_id int
		err = db.QueryRow("SELECT showtime_id FROM showtimes WHERE date=? AND showtime=? AND theater=?", date, showtime, theaterStr).Scan(&showtime_id)
		if err != nil {
			fmt.Println("Error getting showtime ID:", err)
			return
		}

		// Create a slice of integers for seat numbers
		seatNumbers := make([]int, 10)
		for i := 0; i < 10; i++ {
			seatNumbers[i] = i + 1
		}

		// Create a slice of strings for rows
		rows := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

		// Generate a matrix of seat numbers and rows
		var seatList []string
		for _, row := range rows {
			for _, seatNumber := range seatNumbers {
				seatList = append(seatList, fmt.Sprintf("%s%d", row, seatNumber))
			}
		}

		for _, seat := range seatList {
			// Use prepared statement to insert tickets into database
			stmt, err := db.Prepare(`INSERT INTO tickets (showtime_id, date, showtime, title, theater, row, seat, price)
									VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer stmt.Close()

			// Extract the row and seat values from the seat string
			row := string(seat[0])
			seatNumber, _ := strconv.Atoi(seat[1:])

			_, err = stmt.Exec(showtime_id, date, showtime, title, theaterStr, row, seatNumber, price)
			if err != nil {
				fmt.Println("Error inserting showtime:", err)
				return
			}
		}

	}

	fmt.Printf("\n...%d showings added for %s...\n", n_showings, title)
}

func view_showtimes(db *sql.DB) {
	fmt.Println("\n...Viewing Showtimes...")

	// Fetch all the distinct dates from the showtimes table
	rows, err := db.Query("SELECT DISTINCT DATE(date) FROM showtimes")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the dates
	for rows.Next() {
		var date string
		err := rows.Scan(&date)
		if err != nil {
			log.Fatal(err)
		}

		// Fetch the showtimes for the current date
		showtimes, err := db.Query("SELECT * FROM showtimes WHERE DATE(date) = ?", date)
		if err != nil {
			log.Fatal(err)
		}
		defer showtimes.Close()

		// Print the date
		fmt.Printf("Showtimes for %s:\n", date)

		// Iterate over the showtimes
		for showtimes.Next() {
			var showtime_id int
			var title string
			var date string
			var showtime string
			var theater string
			var price float64
			var admin_id int

			err := showtimes.Scan(&showtime_id, &title, &date, &showtime, &theater, &price, &admin_id)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ID: %d, Title: %s, Time: %s, Theater: %s, Price: $%0.2f\n", showtime_id, title, showtime, theater, price)
		}

		// Check for errors while iterating over showtimes
		err = showtimes.Err()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println() // Print a blank line to separate the showtimes for each date
	}

	// Check for errors while iterating over dates
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func buy_tickets(db *sql.DB) {
	var showtime_id, admin_id int
	var input, confirmation, title, date, showtime, theater string
	var price float64

	for {
		fmt.Println("\n-----Ticket Purchasing UI-----")
		fmt.Println("(type 's' for showtimes)")
		fmt.Println("(type 'q' for homepage)")
		fmt.Print("Enter Desired Showtime ID: ")
		fmt.Scanln(&input)
		if input == "q" {
			break
		}
		if input == "s" {
			view_showtimes(db)
			continue
		}
		var err error
		showtime_id, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input, please enter a valid showtime ID")
			continue
		}

		showtime_info, err := db.Query("SELECT * FROM showtimes WHERE showtime_id=?", showtime_id)
		if err != nil {
			log.Fatal(err)
			continue
		}
		defer showtime_info.Close()

		if !showtime_info.Next() {
			log.Fatal("Invalid showtime ID")
			continue
		}

		err = showtime_info.Scan(&showtime_id, &title, &date, &showtime, &theater, &price, &admin_id)
		if err != nil {
			log.Fatal(err)
			continue
		}

		for {
			fmt.Println("\nYou selected this showtime: ")
			fmt.Printf("---> ID: %d, Title: %s, Date: %s, Time: %s, Theater: %s, Price: $%0.2f <---\n", showtime_id, title, date[0:10], showtime, theater, price)
			fmt.Print("Is this showtime correct (y/n)? ")
			fmt.Scanln(&confirmation)
			if confirmation == "y" || confirmation == "Y" {
				tickets, err := db.Query("SELECT row, seat, available FROM tickets WHERE showtime_id=?", showtime_id)
				if err != nil {
					log.Fatal(err)
				}
				defer tickets.Close()

				// Create a 2D slice to hold the ticket information
				var ticketInfo [10][10]string

				// Iterate through the query results and assign the values to the matrix
				for tickets.Next() {
					var row string
					var seat int
					var available string
					err := tickets.Scan(&row, &seat, &available)
					if err != nil {
						log.Fatal(err)
					}

					// Convert the availability string to a string representation
					var availableStr string
					if available == "yes" {
						availableStr = "O"
					} else {
						availableStr = "X"
					}

					// Convert the seat number to an integer index
					rowIdx := int(row[0]) - int('A')
					col := seat - 1

					// Assign the ticket information to the matrix
					ticketInfo[rowIdx][col] = availableStr
				}

				fmt.Println("\n-----Available Tickets-----")
				fmt.Println("(O = Available / X = Taken)")
				fmt.Println("\n  -------SCREEN-------")
				// Print the row numbers
				fmt.Print("  ")
				for i := 1; i <= 10; i++ {
					fmt.Printf("%d ", i)
				}
				fmt.Println()
				// Print the ticket information matrix with row letters and seat numbers
				for i := 0; i < 10; i++ {
					fmt.Printf("%c ", 'A'+i)
					for j := 0; j < 10; j++ {
						fmt.Print(ticketInfo[i][j])
						fmt.Print(" ")
					}
					fmt.Println()
				}

				fmt.Print("\nHow many tickets? ")
				var numTickets int
				fmt.Scanln(&numTickets)

				ticket_price, err := db.Query("SELECT price FROM tickets WHERE showtime_id=?", showtime_id)
				if err != nil {
					log.Fatal(err)
				}
				defer ticket_price.Close()

				var price float64
				if ticket_price.Next() {
					err = ticket_price.Scan(&price)
					if err != nil {
						log.Fatal(err)
					}
				}

				totalPrice := price * float64(numTickets)

				fmt.Printf("Total price for %d tickets: $%.2f\n", numTickets, totalPrice)
				fmt.Print("Do you accept and confirm this price (y/n)? ")
				var confirmation2 string
				fmt.Scanln(&confirmation2)

				if confirmation2 == "y" || confirmation2 == "Y" {
					fmt.Println("Thank you for your purchase!")
					fmt.Println("Continuing to Seat/Row UI")

					var row string
					for {
						fmt.Print("Which Row (A-J)? ")
						_, err := fmt.Scanln(&row)
						if err != nil {
							fmt.Println("Error reading input:", err)
							continue
						}
						if row < "A" || row > "J" {
							fmt.Println("Invalid input. Please enter a letter from A to J.")
							continue
						}
						fmt.Printf("You selected row %s.\n", row)

						for i := 1; i <= numTickets; i++ {
							fmt.Printf("\nTicket %d seat #: ", i)
							var seatNum int
							_, err := fmt.Scanln(&seatNum)
							if err != nil {
								fmt.Println("Error reading input:", err)
								continue
							}
							fmt.Printf("You selected seat %d in row %s.\n", seatNum, row)
						}
						return
					}
				}

			} else if confirmation == "n" || confirmation == "N" {
				fmt.Print("Sorry to hear that. Would you like to try another showtime (y/n)? ")
				fmt.Scanln(&confirmation)
				if confirmation == "y" || confirmation == "Y" {
					buy_tickets(db)
				} else if confirmation == "n" || confirmation == "N" {
					fmt.Println("Returning to homepage.")
					return
				} else {
					fmt.Println("Invalid input. Please enter 'y' or 'n'.")
					continue
				}
			} else {
				fmt.Println("Invalid input. Please enter 'y' or 'n'.")
				continue
			}
			break
		}
		break
	}
}

// Main Function

func main() {
	// Connect to SQLite database
	db, err := sql.Open("sqlite3", "./movie.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create showtimes table
	_, err = db.Exec(createShowtimesTable)
	if err != nil {
		log.Fatal(err)
	}

	// Create tickets table
	_, err = db.Exec(createTicketsTable)
	if err != nil {
		log.Fatal(err)
	}

	// Create Admin table
	_, err = db.Exec(createAdminsTable)
	if err != nil {
		log.Fatal(err)
	}

	// Check if admin account already exists, otherwise create initial admin account
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		_, err = db.Exec("INSERT INTO admins (username, password) VALUES (?, ?)", "admin", "Password1!")
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		// Prompt user for input
		var option int
		fmt.Println("\nGo Movie Ticketing App")
		fmt.Println("Please select an option:")
		fmt.Println("1. View Showtimes")
		fmt.Println("2. Buy Tickets")
		fmt.Println("3. Admin Duties")
		fmt.Println("4. Quit")
		fmt.Print("Option: ")
		fmt.Scanln(&option)

		switch option {
		case 1:
			view_showtimes(db)
		case 2:
			buy_tickets(db)
		case 3:
			// Check if the user is an admin
			adminID := admin_login(db)
			if adminID != 0 {
				admin_abilities(db, adminID)
			}
		case 4:
			fmt.Println("Exiting Application...")
			return // Exit the main function
		default:
			fmt.Println("Invalid option selected")
		}
	}
}
