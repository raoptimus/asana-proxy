package asana

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

type (
	willTask struct {
		Gid  string `json:"id"`
		Name string `json:"name"`
		Key  string `json:"key"`
	}
	gotTaskData struct {
		Gid      string `json:"gid"`
		Id       string `json:"id"`
		Assignee struct {
			Gid          string `json:"gid"`
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
		} `json:"assignee"`
		AssigneeStatus string      `json:"assignee_status"`
		Completed      bool        `json:"completed"`
		CompletedAt    interface{} `json:"completed_at"`
		CreatedAt      time.Time   `json:"created_at"`
		CustomFields   []struct {
			Gid         string `json:"gid"`
			Enabled     bool   `json:"enabled"`
			Name        string `json:"name"`
			NumberValue *int   `json:"number_value,omitempty"`
			Precision   int    `json:"precision,omitempty"`
			CreatedBy   *struct {
				Gid          string `json:"gid"`
				Name         string `json:"name"`
				ResourceType string `json:"resource_type"`
			} `json:"created_by"`
			DisplayValue    *string `json:"display_value"`
			ResourceSubtype string  `json:"resource_subtype"`
			ResourceType    string  `json:"resource_type"`
			Type            string  `json:"type"`
			EnumOptions     []struct {
				Gid          string `json:"gid"`
				Color        string `json:"color"`
				Enabled      bool   `json:"enabled"`
				Name         string `json:"name"`
				ResourceType string `json:"resource_type"`
			} `json:"enum_options,omitempty"`
			EnumValue *struct {
				Gid          string `json:"gid"`
				Color        string `json:"color"`
				Enabled      bool   `json:"enabled"`
				Name         string `json:"name"`
				ResourceType string `json:"resource_type"`
			} `json:"enum_value,omitempty"`
			TextValue *string `json:"text_value,omitempty"`
		} `json:"custom_fields"`
		DueAt     interface{} `json:"due_at"`
		DueOn     string      `json:"due_on"`
		Followers []struct {
			Gid          string `json:"gid"`
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
		} `json:"followers"`
		Hearted     bool          `json:"hearted"`
		Hearts      []interface{} `json:"hearts"`
		Liked       bool          `json:"liked"`
		Likes       []interface{} `json:"likes"`
		Memberships []struct {
			Project struct {
				Gid          string `json:"gid"`
				Name         string `json:"name"`
				ResourceType string `json:"resource_type"`
			} `json:"project"`
			Section struct {
				Gid          string `json:"gid"`
				Name         string `json:"name"`
				ResourceType string `json:"resource_type"`
			} `json:"section"`
		} `json:"memberships"`
		ModifiedAt   time.Time   `json:"modified_at"`
		Name         string      `json:"name"`
		Notes        string      `json:"notes"`
		NumHearts    int         `json:"num_hearts"`
		NumLikes     int         `json:"num_likes"`
		Parent       interface{} `json:"parent"`
		PermalinkUrl string      `json:"permalink_url"`
		Projects     []struct {
			Gid          string `json:"gid"`
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
		} `json:"projects"`
		ResourceType    string        `json:"resource_type"`
		StartAt         interface{}   `json:"start_at"`
		StartOn         interface{}   `json:"start_on"`
		Tags            []interface{} `json:"tags"`
		ResourceSubtype string        `json:"resource_subtype"`
		Workspace       struct {
			Gid          string `json:"gid"`
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
		} `json:"workspace"`
	}
)

var regexpTaskPath *regexp.Regexp
var regexpAuth *regexp.Regexp

func init() {
	var err error
	regexpTaskPath, err = regexp.Compile(".*?/tasks/[0-9]+$")
	if err != nil {
		log.Fatal(err)
	}
	regexpAuth, err = regexp.Compile("Basic\\s+(.*)")
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Proxy) background() {
	log.Infof("server listening %s", s.options.ServerAddr)
	log.Fatal(http.ListenAndServe(s.options.ServerAddr, s))
}

func (s *Proxy) backgroundCacheClear() {
	timer := time.NewTimer(1 * time.Minute)
	for {
		select {
		case <-timer.C:
			s.Lock()
			s.cache = make(map[string]ResponseData)
			s.Unlock()
		}
	}
}

func (s *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := http.DefaultClient

	req := s.createRequest(r)
	cacheKey := fmt.Sprintf("%s: %s", req.Method, req.URL.String())

	s.Lock()
	defer s.Unlock()

	if respData, ok := s.cache[cacheKey]; ok {
		log.Infoln(cacheKey)
		s.writeGotResponseData(w, respData)
	}

	resp, err := client.Do(req)
	if err != nil {
		s.writeServerError(w, err)
		return
	}

	if resp == nil {
		return
	}

	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		s.writeServerError(w, err)
		return
	}

	s.cache[cacheKey] = ResponseData{
		url:        *req.URL,
		data:       data,
		headers:    req.Header,
		statusCode: resp.StatusCode,
	}

	s.writeGotResponseData(w, s.cache[cacheKey])
}

func (s *Proxy) createRequest(r *http.Request) *http.Request {
	baseUrl, _ := url.Parse(s.options.URL)
	reqHeaders := r.Header.Clone()

	authHeader := r.Header.Get("Authorization")
	reqHeaders.Set("Authorization", authHeader)

	if authHeader != "" {
		authMatchGroups := regexpAuth.FindStringSubmatch(authHeader)
		if len(authMatchGroups) > 0 {
			bearer := authMatchGroups[1]
			bearerBase64Data, err := base64.StdEncoding.DecodeString(bearer)
			if err != nil {
				log.Errorf("error decoding authorization header value: %v\n", err)
			} else {
				bearerData := strings.Split(string(bearerBase64Data), ":")
				username, password := bearerData[0], bearerData[1]
				reqHeaders.Set("Authorization", fmt.Sprintf("Bearer 1/%s:%s", username, password))
			}
		}
	}

	return &http.Request{
		Method: r.Method,
		URL: &url.URL{
			Scheme:     baseUrl.Scheme,
			Host:       baseUrl.Host,
			Path:       path.Join(baseUrl.Path, r.URL.Path),
			RawQuery:   r.URL.RawQuery,
			ForceQuery: r.URL.ForceQuery,
		},
		Header: reqHeaders,
	}
}

func (s *Proxy) writeGotResponseData(w http.ResponseWriter, respData ResponseData) {
	for name, values := range respData.headers {
		for _, value := range values {
			log.Debugf("%s: %s\n", name, value)
			w.Header().Add(name, value)
		}
	}

	data := respData.data
	if respData.statusCode == http.StatusOK {
		switch {
		case strings.HasSuffix(respData.url.Path, "/tasks"):
			data = s.replaceResponseTasks(data)
		case regexpTaskPath.Match([]byte(respData.url.Path)):
			data = s.replaceResponseTask(data)
		}
	}

	if _, err := w.Write(data); err != nil {
		log.Error(err)
	}
}

func (s *Proxy) replaceResponseTask(data []byte) []byte {
	gotTask := struct {
		Data gotTaskData `json:"data"`
	}{}

	if err := json.Unmarshal(data, &gotTask); err != nil {
		log.Errorf("cannot decode task: %s\n %s", err, string(data))
		return data
	}

	for _, f := range gotTask.Data.CustomFields {
		if f.TextValue == nil || !f.Enabled || f.Name != "Task short number" {
			continue
		}
		gotTask.Data.Id = *f.TextValue
		break
	}

	replacedResponse, err := json.Marshal(gotTask)
	if err != nil {
		log.Errorf("cannot encode task: %s\n %s", err, string(data))
		return data
	}

	log.Debugf("%s", string(replacedResponse))

	return replacedResponse
}

func (s *Proxy) replaceResponseTasks(data []byte) []byte {
	gotTasks := struct {
		Data []struct {
			Gid          string `json:"gid"`
			Name         string `json:"name"`
			CustomFields []struct {
				Gid         string `json:"gid"`
				Name        string `json:"name"`
				Enabled     bool   `json:"enabled"`
				EnumOptions []struct {
					Gid          string `json:"gid"`
					Color        string `json:"color"`
					Enabled      bool   `json:"enabled"`
					Name         string `json:"name"`
					ResourceType string `json:"resource_type"`
				} `json:"enum_options,omitempty"`
				EnumValue *struct {
					Gid          string `json:"gid"`
					Color        string `json:"color"`
					Enabled      bool   `json:"enabled"`
					Name         string `json:"name"`
					ResourceType string `json:"resource_type"`
				} `json:"enum_value,omitempty"`
				CreatedBy *struct {
					Gid          string `json:"gid"`
					Name         string `json:"name"`
					ResourceType string `json:"resource_type"`
				} `json:"created_by"`
				DisplayValue    *string `json:"display_value"`
				ResourceSubtype string  `json:"resource_subtype"`
				ResourceType    string  `json:"resource_type"`
				Type            string  `json:"type"`
				NumberValue     *int    `json:"number_value,omitempty"`
				Precision       int     `json:"precision,omitempty"`
				TextValue       *string `json:"text_value,omitempty"`
			} `json:"custom_fields"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(data, &gotTasks); err != nil {
		log.Errorf("cannot decode tasks: %s\n %s", err, string(data))
		return data
	}

	willTasks := struct {
		Data []willTask `json:"data"`
	}{}
	willTasks.Data = make([]willTask, 0)

	for _, t := range gotTasks.Data {
		for _, f := range t.CustomFields {
			if f.TextValue == nil || !f.Enabled || f.Name != "Task short number" {
				continue
			}

			willTasks.Data = append(willTasks.Data,
				willTask{
					Gid:  t.Gid,
					Name: t.Name,
					Key:  *f.TextValue,
				},
			)
			break
		}
	}

	replacedData, err := json.Marshal(willTasks)
	if err != nil {
		log.Errorf("cannot decode tasks: %s\n %s", err, string(data))
		return data
	}

	return replacedData
}

func (s *Proxy) writeServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	if _, err := w.Write([]byte(err.Error())); err != nil {
		log.Error(err)
	}
}
