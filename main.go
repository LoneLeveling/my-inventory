package main

func main() {
	app := App{} //created app variable using App struct

	// if err :=
	app.Initialise(DBUser, DbPassword, DBName) /*err != nil {
		log.Fatal(err)
	}*/
	app.Run("localhost:2000")
}
