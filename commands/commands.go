package commands

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Cmolloy36/blog_aggregator/internal/config"
	"github.com/Cmolloy36/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	Db           *database.Queries
	ConfigStruct *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	FunctionMap map[string]func(*State, Command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	slc, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed

	if err = xml.Unmarshal(slc, &rssFeed); err != nil {
		return nil, err
	}

	cleanResHTML(&rssFeed)

	return &rssFeed, nil

}

func cleanResHTML(rssFeed *RSSFeed) {
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for i := range rssFeed.Channel.Item {
		// fmt.Printf("Before: %s\n", item.Title)
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		// fmt.Printf("After:  %s\n", item.Title)

		// fmt.Printf("Before: %s\n", item.Description)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
		// fmt.Printf("After:  %s\n", item.Description)
	}

}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		currentUserName := s.ConfigStruct.Current_user_name

		numRecords, err := s.Db.GetNumRecords(context.Background())
		if err != nil {
			return fmt.Errorf("unexpected error occurred: %v", err)
		} else if numRecords == 0 {
			return fmt.Errorf("no users have been registered")
		}

		user, err := s.Db.GetUser(context.Background(), currentUserName)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unexpected error occurred: %v", err)
			} else {
				return fmt.Errorf("%s does not exist", currentUserName)
			}
		}

		return handler(s, cmd, user)
	}

}

func scrapeFeeds(s *State) error {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("unexpected error occurred in scrapeFeeds: %v", err)
	}

	// When updating the timestamp
	var lastFetchedAt sql.NullTime
	lastFetchedAt.Time = time.Now()
	lastFetchedAt.Valid = true

	markFeedFetchedParams := database.MarkFeedFetchedParams{
		ID:            feed.ID,
		LastFetchedAt: lastFetchedAt,
	}

	err = s.Db.MarkFeedFetched(context.Background(), markFeedFetchedParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred in scrapeFeeds: %v", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("unexpected error occurred in scrapeFeeds: %v", err)
	}

	formats := []string{
		time.RFC1123Z,         // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,          // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
		"2006-01-02 15:04:05", // Custom format
		"2006-01-02",          // Just date
	}

	for i, item := range rssFeed.Channel.Item {
		var title sql.NullString
		title.String = item.Title
		title.Valid = true

		var url sql.NullString
		url.String = item.Link
		url.Valid = true

		var description sql.NullString
		description.String = item.Description
		description.Valid = true

		var publishedAt time.Time

		for _, format := range formats {
			publishedAt, err = time.Parse(format, item.PubDate)
			if err == nil {
				break
			}
		}

		createPostParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   publishedAt,
			Title:       title,
			Url:         url,
			Description: description,
			FeedID:      feed.ID,
		}

		s.Db.CreatePost(context.Background(), createPostParams)

		fmt.Printf("Item %d Title: %s\n", i, item.Title)
		fmt.Printf("Item %d PubDate: %s, PubDate in time: %v\n\n", i, item.PubDate, publishedAt)
	}

	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.FunctionMap[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	fcn, ok := c.FunctionMap[cmd.Name]
	if !ok {
		return fmt.Errorf("error: \"%s\" is not registered", cmd.Name)
	}

	err := fcn(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("error: \"addfeed\" expects feed name & feed url")
	}

	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	createFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID,
	}

	feed, err := s.Db.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred: %v", err)
	}

	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    createFeedParams.ID,
	}

	_, err = s.Db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred in addFeed CreateFeedFollow: %v", err)
	}

	fmt.Printf("%+v", feed)

	return nil
}

func HandlerAggregator(s *State, cmd Command) error {
	var time_between_requests time.Duration
	var err error

	if len(cmd.Args) == 0 {
		time_between_requests, _ = time.ParseDuration("5s")
	} else if len(cmd.Args) == 1 {
		time_between_requests, err = time.ParseDuration(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("unexpected error occurred in HandlerAggregator: %v", err)
		}
	} else {
		return fmt.Errorf("error: \"agg\" expects a no arguments or a time argument (1h, 2m, etc.)")
	}

	fmt.Printf("Collecting feeds every %v\n", time_between_requests)

	ticker := time.NewTicker(time_between_requests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func HandlerBrowser(s *State, cmd Command, user database.User) error {
	limit := 2
	var err error

	// fmt.Printf("%d\n", len(cmd.Args))
	// fmt.Printf("%v\n", reflect.TypeOf(cmd.Args[0]).Kind() == reflect.Int)
	// fmt.Printf("%v\n", reflect.TypeOf(cmd.Args[0]).Kind())

	if len(cmd.Args) == 1 {
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("unexpected error occurred when parsing limit arg: %v", err)
		}
	} else {
		return fmt.Errorf("error: \"browse\" expects a no arguments or an int limit argument")
	}

	getPostsForUserParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.Db.GetPostsForUser(context.Background(), getPostsForUserParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred in HandlerBrowser: %v", err)
	}

	for i, post := range posts {
		fmt.Printf("Post %d: %+v", i, post)
	}

	return nil

}

func HandlerFeeds(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("error: \"feeds\" does not expect an additional argument")
	}

	feedsList, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(feedsList) == 0 {
		return fmt.Errorf("there are no feeds in the database")
	}

	for _, feed := range feedsList {
		username, err := s.Db.GetFeedUser(context.Background(), feed.Url)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		fmt.Printf("Feed Name: %s\n", feed.Name)
		fmt.Printf("Feed url: %s\n", feed.Url)
		fmt.Printf("Feed user: %s\n\n", username)
	}

	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("error: \"follow\" expects a url argument")
	}

	feedURL := cmd.Args[0]

	emptyFeed := database.Feed{}

	feed, err := s.Db.GetFeed(context.Background(), feedURL)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unexpected error occurred: %v", err)
		}
	} else if feed == emptyFeed {
		return fmt.Errorf("feed at %s does not exist", feedURL)
	}

	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.Db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred: %v", err)
	}

	fmt.Printf("%s is now following feed \"%s\"\n", user.Name, feed.Name)

	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("error: \"following\" does not expect any arguments")
	}

	followedFeedList, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unexpected error occurred: %v", err)
		}
	}

	if len(followedFeedList) == 0 {
		fmt.Printf("%s is not following any feeds", user.Name)
		return nil
	}

	fmt.Printf("%s is following:\n", user.Name)

	for _, followedFeed := range followedFeedList {
		fmt.Printf("%s\n", followedFeed.FeedName)
	}

	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("error: \"login\" expects a username argument")
	}

	name := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unexpected error occurred: %v", err)
		} else {
			return fmt.Errorf("%s does not exist", name)
		}
	}

	s.ConfigStruct.Current_user_name = name
	// fmt.Printf("%v", s.ConfigStruct.Current_user_name)
	fmt.Printf("The user has been set: %s\n", s.ConfigStruct.Current_user_name)
	s.ConfigStruct.SetUser(s.ConfigStruct.Current_user_name)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("error: \"register\" expects a username argument")
	}

	name := cmd.Args[0]

	emptyUser := database.User{}

	user, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unexpected error occurred: %v", err)
		}
	} else if user != emptyUser {
		return fmt.Errorf("%s already exists", name)
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		Name:      name,
	}

	_, err = s.Db.CreateUser(context.Background(), userParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred: %v", err)
	}

	s.ConfigStruct.Current_user_name = cmd.Args[0]
	// fmt.Printf("%v", s.ConfigStruct.Current_user_name)
	fmt.Printf("The user has been registered: %s\n", s.ConfigStruct.Current_user_name)
	s.ConfigStruct.SetUser(s.ConfigStruct.Current_user_name)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("unexpected error occurred: %v", err)
	}

	s.ConfigStruct.SetUser("")

	fmt.Println("The database has been reset.")

	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("error: \"unfollow\" expects a url argument")
	}

	feedURL := cmd.Args[0]

	emptyFeed := database.Feed{}

	feed, err := s.Db.GetFeed(context.Background(), feedURL)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unexpected error occurred: %v", err)
		}
	} else if feed == emptyFeed {
		return fmt.Errorf("feed at %s does not exist", feedURL)
	}

	unfollowFeedParams := database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.Db.UnfollowFeed(context.Background(), unfollowFeedParams)
	if err != nil {
		return fmt.Errorf("unexpected error occurred: %v", err)
	}

	fmt.Printf("%s is no longer following feed \"%s\"\n", user.Name, feed.Name)

	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("error: \"users\" does not expect an additional argument")
	}

	usersList, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(usersList) == 0 {
		return fmt.Errorf("there are no users in the database")
	}

	for _, user := range usersList {
		append := ""
		if user == s.ConfigStruct.Current_user_name {
			append = " (current)"
		}
		fmt.Printf("* %s\n", user+append)
	}

	return nil
}
