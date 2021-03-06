package models

import "time"

type CredentialModel struct {
	TableName struct{} `sql:"credentials"`

	ID                      string    	`json:"id" sql:",pk:notnull"`
	ApiKey                  string    	`json:"api_key" sql:",notnull"`
	IPRestriction 			string    	`json:"ip_restriction" sql:",null"`
	HTTPReferrerRestriction string    	`json:"http_referrer_restriction" sql:",null"`
	IOSBundleID				string    	`json:"ios_bundle_restriction" sql:",null"`
	AndroidPackageName		string		`json:"android_package_name" sql:",null"`
	DateCreated             time.Time 	`json:"date_created" sql:",notnull,default:now()"`
}
