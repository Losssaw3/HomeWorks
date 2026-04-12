package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	req              SearchRequest
	ExpectedCount    int
	ExpectedError    error
	ExpectedNextPage bool
}

type tmpUserXML struct {
	Id        string `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       string `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func SortByOrder(arr []User, order int, lessFunc func(i, j int) bool) string {
	switch order {
	case -1:
		sort.Slice(arr, lessFunc)
	case 1:
		sort.Slice(arr, func(i, j int) bool {
			return !lessFunc(i, j)
		})
	case 0:

	default:
		return ErrorBadOrderField
	}
	return ""
}

func SortByFieldWithOrder(arr []User, field string, order int) string {
	var err string = ""
	switch field {
	case "Id":
		err = SortByOrder(arr, order, func(i, j int) bool {
			return arr[i].Id < arr[j].Id
		})
	case "Age":
		err = SortByOrder(arr, order, func(i, j int) bool {
			return arr[i].Age < arr[j].Age
		})
	case "Name":
		err = SortByOrder(arr, order, func(i, j int) bool {
			return arr[i].Name < arr[j].Name
		})
	default:
		err = "ErrorBadOrderField"
	}
	if err != "" {
		return err
	}
	return ""

}

func getInfo(path string) []User {
	storage := []User{}
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "row" {
				var person tmpUserXML
				err = decoder.DecodeElement(&person, &t)
				if err != nil {
					panic(err)
				}
				user := ParseUser(&person)
				storage = append(storage, user)
			}

		}

	}
	return storage
}

func ParseUser(user *tmpUserXML) User {
	ID, _ := strconv.Atoi(user.Id)
	Age, _ := strconv.Atoi(user.Age)
	result := User{
		Id:     ID,
		Name:   strings.TrimSpace(user.FirstName + " " + user.LastName),
		Age:    Age,
		About:  strings.TrimSpace(user.About),
		Gender: user.Gender,
	}
	return result
}

func SearchServerError(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("AccessToken") != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	storage := getInfo("dataset.xml")
	result := []User{}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))

	if limit == 12 { // 12 - trigger for http.StatusInternalServerError
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if query == "" {
		for _, user := range storage {
			result = append(result, user)
		}
	} else {
		for _, user := range storage {
			if strings.Contains(user.About, query) || strings.Contains(user.Name, query) {
				result = append(result, user)

			}
		}
	}

	err := SortByFieldWithOrder(result, orderField, orderBy)
	if err != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{ bad json }"))
		return
	}

	totalFound := len(result)
	start := min(offset, totalFound)

	end := start + limit
	if end < totalFound {
	} else {
		end = totalFound
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{ bad json }"))
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("AccessToken") != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	storage := getInfo("dataset.xml")
	result := []User{}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))

	if query == "" {
		for _, user := range storage {
			result = append(result, user)
		}
	} else {
		for _, user := range storage {
			if strings.Contains(user.About, query) || strings.Contains(user.Name, query) {
				result = append(result, user)

			}
		}
	}

	err := SortByFieldWithOrder(result, orderField, orderBy)
	if err != "" {
		responeErr := &SearchErrorResponse{}
		responeErr.Error = err
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responeErr)
		return
	}

	totalFound := len(result)
	start := min(offset, totalFound)

	end := start + limit
	if end < totalFound {
	} else {
		end = totalFound
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result[start:end])
}

func TestValidRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	cases := []TestCase{
		TestCase{
			req: SearchRequest{Limit: 3,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    3,
			ExpectedError:    nil,
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 25,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    25,
			ExpectedError:    nil,
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 10,
				Offset:     0,
				Query:      "Boyd Wolf",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    1,
			ExpectedError:    nil,
			ExpectedNextPage: false,
		},
	}
	for caseNum, item := range cases {
		srv := &SearchClient{
			AccessToken: "secret",
			URL:         ts.URL,
		}
		result, err := srv.FindUsers(item.req)

		if item.ExpectedError != nil {
			if err == nil {
				t.Errorf("[%d] expected error, got nil", caseNum)
			}
		}

		if err != nil {
			t.Errorf("[%d] unexpected error: %v", caseNum, err)
			continue
		}

		if len(result.Users) != item.ExpectedCount {
			t.Errorf("[%d] wrong users count: expected %d, got %d",
				caseNum, item.ExpectedCount, len(result.Users))
		}

		if result.NextPage != item.ExpectedNextPage {
			t.Errorf("[%d] wrong NextPage: expected %v, got %v",
				caseNum, item.ExpectedNextPage, result.NextPage)
		}
	}

}

func TestInvalidToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	req := SearchRequest{Limit: 40,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    1}
	srv := &SearchClient{
		AccessToken: "wrong_secret",
		URL:         ts.URL,
	}
	_, err := srv.FindUsers(req)
	if err == nil {
		t.Errorf(" expected Bad AccessToken error, got nil")
	}
}

func TestInvalidRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	cases := []TestCase{
		TestCase{
			req: SearchRequest{Limit: 3,
				Offset:     -1,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    3,
			ExpectedError:    fmt.Errorf("offset must be > 0"),
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: -1,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    1,
			ExpectedError:    fmt.Errorf("limit must be > 0"),
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "Wrong field",
				OrderBy:    1},
			ExpectedCount:    5,
			ExpectedError:    fmt.Errorf("OrderFeld Wrong field invalid"),
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    2},
			ExpectedCount:    5,
			ExpectedError:    fmt.Errorf("OrderField invalid"),
			ExpectedNextPage: true,
		},
	}
	for caseNum, item := range cases {
		srv := &SearchClient{
			AccessToken: "secret",
			URL:         ts.URL,
		}
		_, err := srv.FindUsers(item.req)

		if err == nil {
			t.Errorf("[%d] expected error containing '%s', got nil", caseNum, item.ExpectedError.Error())
		} else if !strings.Contains(err.Error(), item.ExpectedError.Error()) {
			t.Errorf("[%d] wrong error: expected '%s', got '%s'", caseNum, item.ExpectedError.Error(), err.Error())
		}

	}
}

func TestBadServerResponce(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerError))
	defer ts.Close()
	cases := []TestCase{
		TestCase{
			req: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "Gender",
				OrderBy:    2},
			ExpectedCount:    5,
			ExpectedError:    fmt.Errorf("cant unpack error json:"),
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 5,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    5,
			ExpectedError:    fmt.Errorf("cant unpack result json:"),
			ExpectedNextPage: true,
		},
		TestCase{
			req: SearchRequest{Limit: 11,
				Offset:     0,
				Query:      "",
				OrderField: "Age",
				OrderBy:    1},
			ExpectedCount:    11,
			ExpectedError:    fmt.Errorf("SearchServer fatal error"),
			ExpectedNextPage: true,
		},
	}
	for caseNum, item := range cases {
		srv := &SearchClient{
			AccessToken: "secret",
			URL:         ts.URL,
		}
		_, err := srv.FindUsers(item.req)

		if err == nil {
			t.Errorf("[%d] expected error containing '%s', got nil", caseNum, item.ExpectedError.Error())
		} else if !strings.Contains(err.Error(), item.ExpectedError.Error()) {
			t.Errorf("[%d] wrong error: expected '%s', got '%s'", caseNum, item.ExpectedError.Error(), err.Error())
		}

	}

}

func TestTimeoutRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	req := SearchRequest{Limit: 40,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    1}
	srv := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	_, err := srv.FindUsers(req)
	if err == nil || !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("expected timeout error, got %v", err)
	}

}

func TestUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	ts.Close()
	req := SearchRequest{Limit: 40,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    1}
	srv := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	_, err := srv.FindUsers(req)
	if err == nil || !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("expected unknown error, got %v", err)
	}

}
