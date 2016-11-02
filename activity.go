package main

import (
	"time"

	gen "github.com/couchbaselabs/setrack/datagen"
)

// Activity structure and JSON Mapping
type Activity struct {
	Type      string `json:"_type"`
	ID        string `json:"_id"`
	CreatedOn string `json:"_createdON"`
	Level     string `json:"Activity Level"`
	Customer  string `json:"Customer or Event Name"`
	DealSize  string `json:"Deal Size $"`
	DealName  string `json:"Deal or Project Name"`
	DealTemp  string `json:"Deal Temperature"`
	Logo      string `json:"New Logo"`
	NextSteps string `json:"Next Steps"`
	Notes     string `json:"Notes"`
	Region    string `json:"Region"`
	SE        string `json:"SE"`
	SFDCLink  string `json:"SFDC Link"`
	Mobile    string `json:"Mobile"`
	Vertical  string `json:"Vertical"`
}

// Instance of an Activity
type SessionActivity struct {
	Activity Activity
}

// Create method off Activity instance stu
func (a *SessionActivity) Create() (*Activity, error) {
	a.Activity.Type = "Activity"
	a.Activity.CreatedOn = time.Now().Format(time.RFC3339)
	a.Activity.ID = GenUUID()

	_, err := bucket.Upsert(a.Activity.ID, a.Activity, 0)
	if err != nil {
		return nil, err
	}
	return &a.Activity, nil
}

// Retrieve and activity by id
func (a SessionActivity) Retrieve(id string) (*Activity, error) {
	_, err := bucket.Get(id, &a.Activity)
	if err != nil {
		return nil, err
	}
	return &a.Activity, nil
}

// Load generator add activity method
func AddActivity() (bool, error) {

	// Local Activity Struct
	var activity SessionActivity

	// Internal Fields
	activity.Activity.Type = "Activity"
	activity.Activity.CreatedOn = time.Now().Format(time.RFC3339)
	activity.Activity.ID = GenUUID()

	// Externally Visible Fields
	activity.Activity.Customer = gen.SillyName()
	activity.Activity.Level = GenINT(1, 3)
	activity.Activity.SE = gen.Engineer()
	activity.Activity.Region = gen.Region()
	activity.Activity.DealName = activity.Activity.Customer
	activity.Activity.DealSize = GenINT(25000, 1000000)
	activity.Activity.Logo = gen.Logo()
	activity.Activity.DealTemp = gen.Temp()
	activity.Activity.Mobile = gen.Mobile()
	activity.Activity.Notes = gen.Paragraph()
	activity.Activity.NextSteps = gen.NextSteps()
	activity.Activity.Vertical = gen.Vertical()

	_, err := activity.Create()
	if err != nil {
		return false, err
	}
	return true, nil
}
