package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Preference struct {
	Text string `json:"text"`
}

type MovieMeta struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Year       int      `json:"year"`
	Rating     float64  `json:"rating"`
	Genres     []string `json:"genres"`
	Overview   string   `json:"overview"`
	PosterLink string   `json:"poster_link"`
}

type Result struct {
	MM    MovieMeta `json:"mm"`
	Likes int       `json:"likes"`
}

type Client struct {
	baseURL        string
	adminToken     string
	userToken      string
	currentRoom    string
	httpClient     *http.Client
	wsConn         *websocket.Conn
	wsDone         chan struct{}
	scanner        *bufio.Scanner
	votingActive   bool
	waitingForVote chan bool
	inputChan      chan string
}

func (c *Client) SetScanner(scanner *bufio.Scanner) {
	c.scanner = scanner
}

type CreateRoomResponse struct {
	Message  string `json:"message"`
	RoomCode string `json:"room_code"`
}

type ParticipateRequest struct {
	Preference Preference `json:"preference"`
}

type ParticipateResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

type MoviesResponse struct {
	Movies []*MovieMeta `json:"movies"`
	Total  int          `json:"total"`
}

type VotingMoviesResponse struct {
	Movies []*MovieMeta `json:"movies"`
}

type ResultsResponse struct {
	Results []*Result `json:"results"`
}

type VoteRequest struct {
	Reactions map[string]int `json:"reactions"`
}

type AuthRequest struct {
	Code string `json:"code"`
}

type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:        baseURL,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		wsDone:         make(chan struct{}),
		waitingForVote: make(chan bool),
		inputChan:      make(chan string),
	}
}

func (c *Client) makeRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	if c.adminToken != "" {
		req.Header.Set("X-admin-token", c.adminToken)
	}
	if c.userToken != "" {
		req.Header.Set("X-user-token", c.userToken)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["Content-Type"]; !exists && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.httpClient.Do(req)
}

func (c *Client) CreateRoom() error {
	resp, err := c.makeRequestWithoutTokens("POST", "/rooms", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create room: %s - %s", resp.Status, string(body))
	}

	c.userToken = resp.Header.Get("X-user-token")

	var response CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	fmt.Printf("Комната создана! Код: %s\n", response.RoomCode)
	fmt.Printf("Ваш токен: %s\n", c.userToken)
	c.currentRoom = response.RoomCode

	fmt.Print("Введите ваши предпочтения фильмов (например: 'action, sci-fi'): ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	preference := strings.TrimSpace(c.scanner.Text())

	participateReq := ParticipateRequest{
		Preference: Preference{
			Text: preference,
		},
	}

	body, err := json.Marshal(participateReq)
	if err != nil {
		return err
	}

	resp, err = c.makeRequest("POST", fmt.Sprintf("/rooms/%s/participations", response.RoomCode),
		strings.NewReader(string(body)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to join room as owner: %s - %s", resp.Status, string(body))
	}

	if err := c.connectWebSocket(response.RoomCode); err != nil {
		return err
	}

	fmt.Println("WebSocket подключен как владелец комнаты")
	return nil
}

func (c *Client) JoinRoom() error {
	fmt.Print("Введите код комнаты: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	roomCode := strings.TrimSpace(c.scanner.Text())

	fmt.Print("Введите ваши предпочтения (описание желаемого фильма): ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	preference := strings.TrimSpace(c.scanner.Text())

	participateReq := ParticipateRequest{
		Preference: Preference{
			Text: preference,
		},
	}

	body, err := json.Marshal(participateReq)
	if err != nil {
		return err
	}

	resp, err := c.makeRequestWithoutTokens("POST", fmt.Sprintf("/rooms/%s/participations", roomCode),
		strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to join room: %s - %s", resp.Status, string(body))
	}

	c.userToken = resp.Header.Get("X-user-token")
	c.currentRoom = roomCode
	fmt.Printf("Ваш токен: %s\n", c.userToken)

	if err := c.connectWebSocket(roomCode); err != nil {
		return err
	}

	fmt.Println("WebSocket подключен к комнате")
	return nil
}

func (c *Client) connectWebSocket(roomCode string) error {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}

	u := url.URL{
		Scheme:   "ws",
		Host:     baseURL.Host,
		Path:     fmt.Sprintf("/api/v1/ws/rooms/%s", roomCode),
		RawQuery: fmt.Sprintf("token=%s", c.userToken),
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}

	c.wsConn = conn
	fmt.Printf("WebSocket подключен к комнате %s\n", roomCode)

	go c.listenWebSocket()

	return nil
}

func (c *Client) StartVoting() error {
	if c.userToken == "" {
		return fmt.Errorf("сначала присоединитесь к комнате")
	}

	if c.wsConn == nil {
		return fmt.Errorf("WebSocket соединение не установлено")
	}

	event := WSEvent{
		Type: "START_VOTING",
	}

	err := c.wsConn.WriteJSON(event)
	if err != nil {
		return fmt.Errorf("failed to start voting: %v", err)
	}

	fmt.Println("Запрос на начало голосования отправлен! Ожидайте...")
	return nil
}

func (c *Client) startVotingProcess() {
	c.votingActive = true
	defer func() { c.votingActive = false }()

	movies, err := c.getVotingMovies(5)
	if err != nil {
		fmt.Printf("Ошибка получения фильмов: %v\n", err)
		return
	}

	fmt.Println("\nФильмы для голосования:")
	for i, movie := range movies {
		fmt.Printf("%d. %s (%d) - Рейтинг: %.1f\n", i+1, movie.Title, movie.Year, movie.Rating)
		fmt.Printf("   Жанры: %s\n", strings.Join(movie.Genres, ", "))
		fmt.Printf("   Описание: %s\n\n", movie.Overview)
	}

	fmt.Printf("Введите ваши реакции для всех %d фильмов через пробел (0 - не нравится, 1 - нравится):\n", len(movies))
	fmt.Print("Например: 1 0 1 1 0\n")
	fmt.Print("Ваши реакции: ")

	select {
	case reactionsInput := <-c.inputChan:
		reactionsInput = strings.TrimSpace(reactionsInput)

		reactionStrs := strings.Fields(reactionsInput)

		if len(reactionStrs) != len(movies) {
			fmt.Printf("Ошибка: ожидается %d реакций, получено %d. Ввод: '%s'\n", len(movies), len(reactionStrs), reactionsInput)
			return
		}

		reactions := make(map[string]int)

		for i, reactionStr := range reactionStrs {
			reaction, err := strconv.Atoi(reactionStr)
			if err != nil || (reaction != 0 && reaction != 1) {
				fmt.Printf("Неверная реакция для фильма '%s': '%s' (должно быть 0 или 1)\n", movies[i].Title, reactionStr)
				return
			}
			reactions[movies[i].ID] = reaction
		}

		fmt.Println("\nВаш выбор:")
		for i, movie := range movies {
			likeText := "не нравится"
			if reactions[movie.ID] == 1 {
				likeText = "нравится"
			}
			fmt.Printf("   %d. %s - %s\n", i+1, movie.Title, likeText)
		}

		err = c.sendVote(reactions)
		if err != nil {
			fmt.Printf("Ошибка отправки голоса: %v\n", err)
			return
		}

		fmt.Println("\nВаш голос принят! Ожидаем других участников...")

	case <-time.After(30 * time.Second):
		fmt.Println("\nТаймаут: вы не ввели реакции вовремя")
		return
	}

	<-c.waitingForVote
}

func (c *Client) makeRequestWithoutTokens(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) listenWebSocket() {
	defer close(c.wsDone)

	for {
		var event WSEvent
		err := c.wsConn.ReadJSON(&event)
		if err != nil {
			fmt.Printf("WebSocket error: %v\n", err)
			return
		}

		switch event.Type {
		case "LOBBY_UPDATE":
			if payload, ok := event.Payload.(map[string]interface{}); ok {
				if count, exists := payload["participants_count"]; exists {
					fmt.Printf("Обновление лобби: %v участников\n", count)
				}
			}

		case "REDIRECT_TO_VOTING":
			fmt.Println("Голосование начато! Получаем фильмы для голосования...")
			go c.startVotingProcess()

		case "VOTING_FINISHED":
			fmt.Println("Все участники проголосовали! Загружаем результаты...")
			select {
			case c.waitingForVote <- true:
			default:
			}
			c.showResults()

		case "ERROR":
			if payload, ok := event.Payload.(map[string]interface{}); ok {
				if message, exists := payload["message"]; exists {
					fmt.Printf("Ошибка: %v\n", message)
				}
			}
		}
	}
}

func (c *Client) getVotingMovies(count int) ([]*MovieMeta, error) {
	resp, err := c.makeRequest("GET",
		fmt.Sprintf("/rooms/%s/movies?count=%d", c.currentRoom, count), nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get movies: %s - %s", resp.Status, string(body))
	}

	var response VotingMoviesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Movies, nil
}

func (c *Client) sendVote(reactions map[string]int) error {
	voteReq := VoteRequest{
		Reactions: reactions,
	}

	body, err := json.Marshal(voteReq)
	if err != nil {
		return err
	}

	resp, err := c.makeRequest("PATCH",
		fmt.Sprintf("/rooms/%s/results", c.currentRoom),
		strings.NewReader(string(body)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vote failed: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (c *Client) showResults() {
	results, err := c.getResults()
	if err != nil {
		fmt.Printf("Ошибка получения результатов: %v\n", err)
		return
	}

	fmt.Println("\nРезультаты голосования:")
	if len(results) == 0 {
		fmt.Println("Нет результатов для отображения")
		return
	}

	for i, result := range results {
		fmt.Printf("%d. %s (%d) - %d голосов\n",
			i+1, result.MM.Title, result.MM.Year, result.Likes)
		fmt.Printf("   Рейтинг: %.1f | Жанры: %s\n",
			result.MM.Rating, strings.Join(result.MM.Genres, ", "))
		fmt.Printf("   Описание: %s\n\n", result.MM.Overview)
	}
}

func (c *Client) getResults() ([]*Result, error) {
	resp, err := c.makeRequest("GET",
		fmt.Sprintf("/rooms/%s/results", c.currentRoom), nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get results: %s - %s", resp.Status, string(body))
	}

	var response ResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (c *Client) AdminAuth() error {
	fmt.Print("Введите код администратора: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	code := strings.TrimSpace(c.scanner.Text())

	authReq := AuthRequest{Code: code}
	body, err := json.Marshal(authReq)
	if err != nil {
		return err
	}

	resp, err := c.makeRequest("POST", "/auth", strings.NewReader(string(body)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s - %s", resp.Status, string(body))
	}

	c.adminToken = resp.Header.Get("X-admin-token")
	fmt.Println("Аутентификация администратора успешна!")
	return nil
}

func (c *Client) AddMovie() error {
	if c.adminToken == "" {
		return fmt.Errorf("сначала выполните аутентификацию администратора")
	}

	fmt.Print("Название фильма: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	title := strings.TrimSpace(c.scanner.Text())

	fmt.Print("Год выпуска: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	yearStr := strings.TrimSpace(c.scanner.Text())
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return fmt.Errorf("неверный формат года")
	}

	fmt.Print("Рейтинг: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	ratingStr := strings.TrimSpace(c.scanner.Text())
	rating, err := strconv.ParseFloat(ratingStr, 64)
	if err != nil {
		return fmt.Errorf("неверный формат рейтинга")
	}

	fmt.Print("Жанры (через запятую): ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	genresStr := strings.TrimSpace(c.scanner.Text())
	genres := strings.Split(genresStr, ",")
	for i := range genres {
		genres[i] = strings.TrimSpace(genres[i])
	}

	fmt.Print("Описание: ")
	if !c.scanner.Scan() {
		return fmt.Errorf("ошибка чтения ввода")
	}
	overview := strings.TrimSpace(c.scanner.Text())

	movieData := map[string]interface{}{
		"title":    title,
		"year":     year,
		"rating":   rating,
		"genres":   genres,
		"overview": overview,
	}

	jsonData, err := json.Marshal(movieData)
	if err != nil {
		return err
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormField("body")
	if err != nil {
		return err
	}
	part.Write(jsonData)

	err = writer.Close()
	if err != nil {
		return err
	}

	resp, err := c.makeRequest("POST", "/movies",
		&requestBody,
		map[string]string{"Content-Type": writer.FormDataContentType()})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add movie: %s - %s", resp.Status, string(body))
	}

	fmt.Println("Фильм успешно добавлен!")
	return nil
}

func (c *Client) ViewMovies() error {
	if c.adminToken == "" {
		return fmt.Errorf("сначала выполните аутентификацию администратора")
	}

	resp, err := c.makeRequest("GET", "/movies", nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get movies: %s - %s", resp.Status, string(body))
	}

	var response MoviesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	fmt.Printf("\nВсего фильмов: %d\n", response.Total)
	for i, movie := range response.Movies {
		fmt.Printf("\n%d. %s (%d)\n", i+1, movie.Title, movie.Year)
		fmt.Printf("   Рейтинг: %.1f\n", movie.Rating)
		fmt.Printf("   Жанры: %s\n", strings.Join(movie.Genres, ", "))
		fmt.Printf("   Описание: %s\n", movie.Overview)
		if movie.PosterLink != "" {
			fmt.Printf("   Постер: %s\n", movie.PosterLink)
		}
	}

	return nil
}

func (c *Client) Close() {
	if c.wsConn != nil {
		c.wsConn.Close()
		<-c.wsDone
	}
}

func main() {
	client := NewClient("http://localhost:8080/api/v1")
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)
	client.SetScanner(scanner)

	for {
		if !client.votingActive {
			fmt.Println("\n=== Kinoswap Console Client ===")
			fmt.Println("1. Создать комнату")
			fmt.Println("2. Войти в комнату")
			fmt.Println("3. Начать голосование")
			fmt.Println("4. Войти как администратор")
			fmt.Println("5. Добавить фильм")
			fmt.Println("6. Просмотреть фильмы")
			fmt.Println("0. Выход")
			fmt.Print("Выберите действие: ")
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if client.votingActive {
			select {
			case client.inputChan <- input:
			default:
				fmt.Println("Подождите, обрабатывается предыдущий ввод...")
			}
			continue
		}

		switch input {
		case "1":
			if err := client.CreateRoom(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "2":
			if err := client.JoinRoom(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "3":
			if err := client.StartVoting(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "4":
			if err := client.AdminAuth(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "5":
			if err := client.AddMovie(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "6":
			if err := client.ViewMovies(); err != nil {
				fmt.Printf("Ошибка: %v\n", err)
			}
		case "0":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Неверный выбор")
		}

		if !client.votingActive {
			fmt.Println("\nНажмите Enter чтобы продолжить...")
			scanner.Scan()
		}
	}
}
