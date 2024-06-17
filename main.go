package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//router
	router := mux.NewRouter()
	//route handlers
	router.HandleFunc("/product", getProducts(db)).Methods("GET")
	router.HandleFunc("/product/:id", getProductByID(db)).Methods("GET")
	router.HandleFunc("/product", addProduct(db)).Methods("POST")
	router.HandleFunc("/product/:id", updateProduct(db)).Methods(("PUT"))
	router.HandleFunc("/product/:id", deleteProduct(db)).Methods(("DELETE"))

	//start server
	log.Fatal(http.ListenAndServe(":8080", jsonContentTypeMiddleware(router)))
}

// this middleware func is wrapped around the router
// any incomding request first hits the middleware func, and all it does is set the response type to json and then call the "next func"
// which is going to be the func that matches the router path. i.e. a get request to /products would first hit middleware and then call getproducts
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all products
func getProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM products")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		products := []Product{}

		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Quantity); err != nil {
				log.Fatal(err)
			}

			products = append(products, p)
		}

		if err := rows.Err(); err != nil {
			log.Fatal()
		}

		json.NewEncoder(w).Encode(products)
	}
}

// with mux, Vars returns a map. the keys are param names defined in the route and values are what is extracted from the URL.
func getProductByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var p Product

		if err := db.QueryRow("SELECT * FROM products WHERE id = $1", id).Scan(&p.ID, &p.Name, &p.Quantity); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(p)
	}
}

// decode json from body of request into p which is a struct instance and then insert into db
func addProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Product
		json.NewDecoder(r.Body).Decode(&p)

		if err := db.QueryRow("INSERT INTO products (name, quantity) VALUES ($1, $2) RETURNING id", p.Name, p.Quantity).Scan(&p.ID); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(p)
	}
}

//create new product, deserialize json from r aka the request that would contain json in the body, and then input that info into the struct with decode and product reference passed in

func updateProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var p Product
		json.NewDecoder(r.Body).Decode(&p)

		vars := mux.Vars(r)
		id := vars["id"]

		if _, err := db.Exec("UPDATE products SET name =  $1, quantity = $2 WHERE id = $3", p.Name, p.Quantity, id); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(p)
	}
}

func deleteProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		if _, err := db.Exec("DELETE FROM products where id = $1", id); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode("Product Deleted")

	}
}
