test:
	go test -v -race
dbUp:
	docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
run:
	go run main.go db_explorer.go
