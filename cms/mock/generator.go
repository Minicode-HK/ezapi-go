package mockdata

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

var (
	// Expanded first names: Top ~50 mixed from SSA 2023 data
	firstNames = []string{
		"John", "Jane", "Michael", "Sarah", "David", "Emily", "James", "Emma", "Robert", "Olivia",
		"William", "Sophia", "Joseph", "Ava", "Charles", "Isabella", "Liam", "Noah", "Oliver", "Amelia",
		"Theodore", "Charlotte", "Elijah", "Luna", "Henry", "Harper", "Mateo", "Camila", "Sebastian", "Gianna",
		"Lucas", "Evelyn", "Levi", "Abigail", "Benjamin", "Ella", "Jack", "Elizabeth", "Ezra", "Sofia",
		"Daniel", "Avery", "Ethan", "Scarlett", "Jacob", "Grace", "Logan", "Chloe", "Samuel", "Penelope",
	}

	// Expanded last names: Top ~50 from US Census data
	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
		"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson",
		"Walker", "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts",
	}

	// Expanded categories: Common e-commerce product categories
	categories = []string{
		"Electronics", "Clothing", "Food", "Books", "Toys", "Sports", "Home", "Garden", "Automotive", "Health",
		"Beauty", "Music", "Movies", "Software", "Furniture", "Beverages", "Accessories", "Hardware", "Media",
		"Personal Care", "Baby", "Pets", "Jewelry", "Shoes", "Electronics Accessories", "Office Supplies",
	}

	// Expanded cities: Top ~50 largest US cities by population
	cities = []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego",
		"Dallas", "San Jose", "Austin", "Jacksonville", "Fort Worth", "Columbus", "Charlotte", "San Francisco",
		"Indianapolis", "Seattle", "Denver", "Washington", "Boston", "El Paso", "Nashville", "Detroit", "Oklahoma City",
		"Portland", "Las Vegas", "Memphis", "Louisville", "Baltimore", "Milwaukee", "Albuquerque", "Tucson",
		"Fresno", "Sacramento", "Mesa", "Kansas City", "Atlanta", "Long Beach", "Colorado Springs", "Raleigh",
		"Miami", "Virginia Beach", "Omaha", "Oakland", "Minneapolis", "Tulsa", "Arlington", "Tampa", "New Orleans",
		"Wichita",
	}

	// Expanded streets: More common street names with varied prefixes and suffixes
	streets = []string{
		"Main St", "Oak Ave", "Maple Dr", "Cedar Ln", "Pine Rd", "Elm St", "Washington Blvd", "Park Ave",
		"Broadway", "Market St", "High St", "River Rd", "Lake Dr", "Hill Ave", "Valley Blvd", "Forest Ct",
		"Spring Pl", "Summer Way", "Winter Cir", "Autumn Ter", "Rose Sq", "Lily Ln", "First Ave", "Second St",
		"Third Dr", "Central Blvd", "North Park", "South Main", "East Oak", "West Elm", "Green St",
		"Blue Ave", "Red Rd", "White Ln", "King St", "Queen Blvd", "Lincoln Dr", "Jefferson Ave",
	}

	// Expanded domains: Popular email providers
	domains = []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "example.com", "icloud.com", "protonmail.com",
		"aol.com", "comcast.net", "att.net", "verizon.net", "sbcglobal.net", "msn.com", "live.com",
	}

	// Expanded loremWords: More words from extended Lorem Ipsum passages
	loremWords = []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do",
		"eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua", "ut",
		"enim", "ad", "minim", "veniam", "quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi",
		"ut", "aliquip", "ex", "ea", "commodo", "consequat", "duis", "aute", "irure", "dolor",
		"in", "reprehenderit", "in", "voluptate", "velit", "esse", "cillum", "dolore", "eu", "fugiat",
		"nulla", "pariatur", "excepteur", "sint", "occaecat", "cupidatat", "non", "proident", "sunt",
		"in", "culpa", "qui", "officia", "deserunt", "mollit", "anim", "id", "est", "laborum",
		"exercitationem", "ullamcorper", "suscipit", "lobortis", "sapien", "nunc", "bibendum", "dolor", "feugiat",
		"at", "nisl", "nulla", "facilisi", "nunc", "sed", "augue", "lacus", "ultricies", "sed",
	}

	// Expanded countries: Top ~20 most populous
	countries = []string{
		"United States", "Canada", "United Kingdom", "Germany", "France", "Australia", "Japan", "India",
		"Brazil", "Mexico", "China", "Indonesia", "Pakistan", "Nigeria", "Bangladesh", "Russia", "Ethiopia",
		"Philippines", "Egypt", "DR Congo", "Vietnam", "Iran", "Turkey", "Thailand", "Germany",
	}

	// Expanded statuses: Common e-commerce and general app statuses
	statuses = []string{
		"active", "inactive", "pending", "completed", "cancelled", "draft", "archived", "shipped",
		"processing", "failed", "on_hold", "refunded", "partially_shipped", "awaiting_payment", "paid",
		"awaiting_fulfillment", "awaiting_pickup", "awaiting_shipment", "captured", "partially_refunded",
	}

	// Expanded roles: Common web app and system roles
	roles = []string{
		"admin", "user", "moderator", "guest", "editor", "superadmin", "owner", "contributor",
		"viewer", "subscriber", "author", "publisher", "manager", "administrator", "creator",
		"executive", "policy_admin", "account_admin",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Generator struct {
	// Config allows customization (e.g., seed, ranges)
	Config GeneratorConfig
}

type GeneratorConfig struct {
	// Min/Max for various generators
	MinAge, MaxAge          int
	MinPrice, MaxPrice      float64
	MinInt, MaxInt          int
	MinFloat, MaxFloat      float64
	// For slices: default max length
	MaxSliceLength          int
	// Lorem sentence ranges
	MinSentenceWords, MaxSentenceWords int
	// Date range: days back from now
	MaxDateDaysBack int
	// Enable/disable recursion depth for nested structs
	MaxRecursionDepth int
}

func NewGenerator() *Generator {
	return &Generator{
		Config: GeneratorConfig{
			MinAge:               18,
			MaxAge:               80,
			MinPrice:             0.01,
			MaxPrice:             999.99,
			MinInt:               1,
			MaxInt:               10000,
			MinFloat:             0.0,
			MaxFloat:             1000.0,
			MaxSliceLength:       10,
			MinSentenceWords:     5,
			MaxSentenceWords:     20,
			MaxDateDaysBack:      365 * 5, // 5 years back
			MaxRecursionDepth:    5,
		},
	}
}

// GenerateValue generates a mock value based on field name and type
// (kept for backward compatibility; enhanced with config)
func (g *Generator) GenerateValue(fieldName string, fieldType reflect.Type) interface{} {
	lowerName := strings.ToLower(fieldName)
	// Check field name patterns
	switch {
	case lowerName == "id" || strings.Contains(lowerName, "uuid"):
		return g.UUID()
	case strings.Contains(lowerName, "email"):
		return g.Email()
	case strings.Contains(lowerName, "name") && !strings.Contains(lowerName, "username"):
		if strings.Contains(lowerName, "first") {
			return g.FirstName()
		} else if strings.Contains(lowerName, "last") {
			return g.LastName()
		}
		return g.FullName()
	case strings.Contains(lowerName, "username"):
		return g.Username()
	case strings.Contains(lowerName, "password"):
		return g.Password()
	case strings.Contains(lowerName, "price") || strings.Contains(lowerName, "amount") || strings.Contains(lowerName, "cost"):
		return g.Price()
	case strings.Contains(lowerName, "phone") || strings.Contains(lowerName, "mobile"):
		return g.Phone()
	case strings.Contains(lowerName, "address"):
		return g.Address()
	case strings.Contains(lowerName, "street"):
		return g.Street()
	case strings.Contains(lowerName, "city"):
		return g.City()
	case strings.Contains(lowerName, "state"):
		return g.State()
	case strings.Contains(lowerName, "zip") || strings.Contains(lowerName, "postal"):
		return g.ZipCode()
	case strings.Contains(lowerName, "country"):
		return g.Country()
	case strings.Contains(lowerName, "description") || strings.Contains(lowerName, "summary") || strings.Contains(lowerName, "bio"):
		return g.Paragraph(3, 5) // Multiple sentences
	case strings.Contains(lowerName, "title"):
		return g.Title()
	case strings.Contains(lowerName, "url") || strings.Contains(lowerName, "website") || strings.Contains(lowerName, "link"):
		return g.URL()
	case strings.Contains(lowerName, "image") || strings.Contains(lowerName, "avatar"):
		return g.ImageURL()
	case strings.Contains(lowerName, "category"):
		return g.Category()
	case strings.Contains(lowerName, "tag") || strings.Contains(lowerName, "tags"):
		return g.Tags(3, 5)
	case strings.Contains(lowerName, "status"):
		return g.Status()
	case strings.Contains(lowerName, "role"):
		return g.Role()
	case strings.Contains(lowerName, "age"):
		return g.Age()
	case strings.Contains(lowerName, "date") || strings.Contains(lowerName, "createdat") || strings.Contains(lowerName, "updatedat") || strings.Contains(lowerName, "birthday"):
		if fieldType == reflect.TypeOf(time.Time{}) {
			return g.Time()
		}
		return g.Date()
	case strings.Contains(lowerName, "boolean") || strings.Contains(lowerName, "active") || strings.Contains(lowerName, "enabled") || strings.Contains(lowerName, "verified"):
		return g.Boolean()
	case strings.Contains(lowerName, "company") || strings.Contains(lowerName, "organization"):
		return g.Company()
	case strings.Contains(lowerName, "lat") || strings.Contains(lowerName, "longitude"):
		return g.Longitude()
	case strings.Contains(lowerName, "lng") || strings.Contains(lowerName, "latitude"):
		return g.Latitude()
	}

	// Generate based on type
	switch fieldType.Kind() {
	case reflect.String:
		return g.Word()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.IntBetween(g.Config.MinInt, g.Config.MaxInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint64(g.IntBetween(1, int(g.Config.MaxInt)))
	case reflect.Float32, reflect.Float64:
		return g.FloatBetween(g.Config.MinFloat, g.Config.MaxFloat)
	case reflect.Bool:
		return g.Boolean()
	default:
		return nil
	}
}

// Populate populates a struct (or ptr to struct) with mock data recursively.
// Supports nested structs, slices, and maps (basic).
// Returns error if not a struct or ptr to struct.
func (g *Generator) Populate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("populate: expected struct or ptr to struct, got %s", val.Kind())
	}
	return g.populateStruct(val, 1)
}

func (g *Generator) populateStruct(val reflect.Value, depth int) error {
	if depth > g.Config.MaxRecursionDepth {
		return nil // Stop recursion
	}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle different kinds
		switch fieldVal.Kind() {
		case reflect.Struct:
			if fieldVal.CanSet() {
				if err := g.populateStruct(fieldVal, depth+1); err != nil {
					return err
				}
			}
		case reflect.Slice:
			if fieldVal.CanSet() && !fieldVal.IsNil() {
				sliceLen := rand.Intn(g.Config.MaxSliceLength) + 1
				elemType := field.Type.Elem()
				slice := reflect.MakeSlice(field.Type, sliceLen, sliceLen)
				for j := 0; j < sliceLen; j++ {
					elem := reflect.New(elemType).Elem()
					if elemType.Kind() == reflect.Struct {
						if err := g.populateStruct(elem, depth+1); err != nil {
							return err
						}
					} else {
						mockVal := g.GenerateValue(field.Name, elemType)
						elem.Set(reflect.ValueOf(mockVal))
					}
					slice.Index(j).Set(elem)
				}
				fieldVal.Set(slice)
			}
		case reflect.Map:
			if fieldVal.CanSet() && !fieldVal.IsNil() {
				mapType := field.Type
				mapLen := rand.Intn(5) + 1
				m := reflect.MakeMapWithSize(mapType, mapLen)
				keyType := mapType.Key()
				valType := mapType.Elem()
				for j := 0; j < mapLen; j++ {
					key := reflect.New(keyType).Elem()
					key.SetString(g.Word())
					value := reflect.New(valType).Elem()
					mockVal := g.GenerateValue(fmt.Sprintf("map_%d", j), valType)
					value.Set(reflect.ValueOf(mockVal))
					m.SetMapIndex(key, value)
				}
				fieldVal.Set(m)
			}
		default:
			if fieldVal.CanSet() {
				mockVal := g.GenerateValue(field.Name, field.Type)
				fieldVal.Set(reflect.ValueOf(mockVal))
			}
		}
	}
	return nil
}

// Enhanced string generators
func (g *Generator) FirstName() string {
	return firstNames[rand.Intn(len(firstNames))]
}

func (g *Generator) LastName() string {
	return lastNames[rand.Intn(len(lastNames))]
}

func (g *Generator) FullName() string {
	return g.FirstName() + " " + g.LastName()
}

func (g *Generator) Username() string {
	return strings.ToLower(g.FirstName() + g.LastName() + fmt.Sprintf("%d", rand.Intn(999)))
}

func (g *Generator) Password() string {
	// Simple 8-12 char password with mix
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$"
	pass := make([]byte, rand.Intn(5)+8)
	for i := range pass {
		pass[i] = chars[rand.Intn(len(chars))]
	}
	return string(pass)
}

func (g *Generator) Email() string {
	return strings.ToLower(g.FirstName()) + "." + strings.ToLower(g.LastName()) + "@" + domains[rand.Intn(len(domains))]
}

func (g *Generator) Phone() string {
	return fmt.Sprintf("(%03d) %03d-%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(9000)+1000)
}

func (g *Generator) Street() string {
	return fmt.Sprintf("%d %s", rand.Intn(9999)+1, streets[rand.Intn(len(streets))])
}

func (g *Generator) Address() string {
	return g.Street() + ", " + g.City() + ", " + g.State() + " " + g.ZipCode()
}

func (g *Generator) City() string {
	return cities[rand.Intn(len(cities))]
}

func (g *Generator) State() string {
	states := []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
	return states[rand.Intn(len(states))]
}

func (g *Generator) ZipCode() string {
	return fmt.Sprintf("%05d", rand.Intn(99999)+1)
}

func (g *Generator) Country() string {
	return countries[rand.Intn(len(countries))]
}

func (g *Generator) Company() string {
	return g.Word() + " " + g.Word() + " Inc."
}

func (g *Generator) Category() string {
	return categories[rand.Intn(len(categories))]
}

func (g *Generator) Tags(min, max int) []string {
	count := rand.Intn(max-min+1) + min
	tags := make([]string, count)
	for i := 0; i < count; i++ {
		tags[i] = g.Word()
	}
	return tags
}

func (g *Generator) Status() string {
	return statuses[rand.Intn(len(statuses))]
}

func (g *Generator) Role() string {
	return roles[rand.Intn(len(roles))]
}

func (g *Generator) Word() string {
	return loremWords[rand.Intn(len(loremWords))]
}

func (g *Generator) Sentence() string {
	return g.SentenceWith(g.Config.MinSentenceWords, g.Config.MaxSentenceWords)
}

func (g *Generator) SentenceWith(minWords, maxWords int) string {
	wordCount := rand.Intn(maxWords-minWords+1) + minWords
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = g.Word()
	}
	sentence := strings.Join(words, " ")
	return strings.ToUpper(sentence[0:1]) + sentence[1:] + "."
}

func (g *Generator) Paragraph(minSentences, maxSentences int) string {
	sentenceCount := rand.Intn(maxSentences-minSentences+1) + minSentences
	var sentences []string
	for i := 0; i < sentenceCount; i++ {
		sentences = append(sentences, g.Sentence())
	}
	return strings.Join(sentences, " ")
}

func (g *Generator) Title() string {
	words := 3 + rand.Intn(4) // 3-6 words
	titleWords := make([]string, words)
	for i := 0; i < words; i++ {
		titleWords[i] = g.Word()
	}
	title := strings.Join(titleWords, " ")
	return strings.ToUpper(title[0:1]) + title[1:]
}

func (g *Generator) UUID() string {
	// Simple random hex UUID v4 mock (not cryptographically secure, but fine for mocks)
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (g *Generator) URL() string {
	return "https://www." + strings.ToLower(g.Word()) + "." + domains[rand.Intn(len(domains))]
}

func (g *Generator) ImageURL() string {
	width := rand.Intn(400) + 200
	height := rand.Intn(400) + 200
	return fmt.Sprintf("https://picsum.photos/%d/%d", width, height)
}

// Number generators (use config)
func (g *Generator) IntBetween(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func (g *Generator) FloatBetween(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (g *Generator) Price() float64 {
	return g.FloatBetween(g.Config.MinPrice, g.Config.MaxPrice)
}

func (g *Generator) Age() int {
	return g.IntBetween(g.Config.MinAge, g.Config.MaxAge)
}

func (g *Generator) Latitude() float64 {
	return g.FloatBetween(-90.0, 90.0)
}

func (g *Generator) Longitude() float64 {
	return g.FloatBetween(-180.0, 180.0)
}

// Date generators
func (g *Generator) Date() string {
	days := rand.Intn(g.Config.MaxDateDaysBack)
	date := time.Now().AddDate(0, 0, -days)
	return date.Format(time.RFC3339)
}

func (g *Generator) Time() time.Time {
	days := rand.Intn(g.Config.MaxDateDaysBack)
	return time.Now().AddDate(0, 0, -days)
}

// Boolean generator
func (g *Generator) Boolean() bool {
	return rand.Intn(2) == 1
}