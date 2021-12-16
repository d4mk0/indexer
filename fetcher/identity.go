package fetcher

import (
	"encoding/json"
	"fmt"
	"bytes"

	"github.com/valyala/fastjson"

	"go.uber.org/zap"
)

const IdentityApiCount = 3

func (f *fetcher) FetchIdentity(address string) (IdentityEntryList, error) {

	var identityArr IdentityEntryList
	ch := make(chan IdentityEntry)

	// Part 1 - Demo data source
	// Context API
	go f.processContext(address, ch)
	// Superrare API
	go f.processSuperrare(address, ch)
	// Poap API
	go f.processPoap(address, ch)

	// Final Part - Merge entry
	for i := 0; i < IdentityApiCount; i++ {
		entry := <-ch
		if entry.Err != nil {
			zap.L().With(zap.Error(entry.Err)).Error("identity api error: " + entry.Msg)
			continue
		}
		if entry.OpenSea != nil {
			identityArr.OpenSea = append(identityArr.OpenSea, *entry.OpenSea)
		}
		if entry.Twitter != nil {
			entry.Twitter.Handle = convertTwitterHandle(entry.Twitter.Handle)
			identityArr.Twitter = append(identityArr.Twitter, *entry.Twitter)
		}
		if entry.Superrare != nil {
			identityArr.Superrare = append(identityArr.Superrare, *entry.Superrare)
		}
		if entry.Rarible != nil {
			identityArr.Rarible = append(identityArr.Rarible, *entry.Rarible)
		}
		if entry.Context != nil {
			identityArr.Context = append(identityArr.Context, *entry.Context)
		}
		if entry.Zora != nil {
			identityArr.Zora = append(identityArr.Zora, *entry.Zora)
		}
		if entry.Foundation != nil {
			identityArr.Foundation = append(identityArr.Foundation, *entry.Foundation)
		}
		if entry.Showtime != nil {
			identityArr.Showtime = append(identityArr.Showtime, *entry.Showtime)
		}
		if entry.Poap != nil {
			identityArr.Poap = append(identityArr.Poap, entry.Poap...)
		}
		if entry.Ens != nil {
			identityArr.Ens = entry.Ens.Ens
		}
	}

	return identityArr, nil
}

func (f *fetcher) processContext(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(ContextUrl, address),
		method: "GET",
	})
	if err != nil {
		result.Err = err
		result.Msg = "[processContext] fetch identity failed"
		ch <- result
		return
	}
	contextProfile := ContextAppResp{}
	err = json.Unmarshal(body, &contextProfile)
	if err != nil {
		result.Err = err
		result.Msg = "[processContext] identity response json unmarshal failed"
		ch <- result
		return
	}

	if value, ok := contextProfile.Ens[address]; ok {
		result.Ens = &UserEnsIdentity{
			Ens:        value,
			DataSource: CONTEXT,
		}
	}

	for _, profileList := range contextProfile.Profiles {
		for _, entry := range profileList {
			switch entry.Contract {
			case SuperrareContractAddress:
				result.Superrare = &UserSuperrareIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case OpenSeaContractAddress:
				result.OpenSea = &UserOpenSeaIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case RaribleContractAddress:
				result.Rarible = &UserRaribleIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case FoundationContractAddress:
				result.Foundation = &UserFoundationIdentity{
					Website:    entry.Website,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case ZoraContractAddress:
				result.Zora = &UserZoraIdentity{
					Website:    entry.Website,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case ContextContractAddress:
				result.Context = &UserContextIdentity{
					Username:      entry.Username,
					Website:       entry.Website,
					FollowerCount: contextProfile.FollowerCount,
					DataSource:    CONTEXT,
				}
			default:
			}
		}
	}

	ch <- result
	return
}

func (f *fetcher) processSuperrare(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(SuperrareUrl, address),
		method: "GET",
	})
	if err != nil {
		result.Err = err
		result.Msg = "[processSuperrare] fetch identity failed"
		ch <- result
		return
	}

	sprProfile := SuperrareProfile{}
	err = json.Unmarshal(body, &sprProfile)
	if err != nil {
		result.Err = err
		result.Msg = "[processSuperrare] identity response json unmarshal failednti"
		ch <- result
		return
	}

	newSprRecord := UserSuperrareIdentity{
		Username:       sprProfile.Result.Username,
		Location:       sprProfile.Result.Location,
		Bio:            sprProfile.Result.Bio,
		InstagramLink:  sprProfile.Result.InstagramLink,
		TwitterLink:    sprProfile.Result.TwitterLink,
		SteemitLink:    sprProfile.Result.SteemitLink,
		Website:        sprProfile.Result.Website,
		SpotifyLink:    sprProfile.Result.SpotifyLink,
		SoundCloudLink: sprProfile.Result.SoundCloudLink,
		DataSource:     SUPERRARE,
	}

	if newSprRecord.Username != "" || newSprRecord.Location != "" || newSprRecord.Bio != "" || newSprRecord.InstagramLink != "" ||
		newSprRecord.TwitterLink != "" || newSprRecord.SteemitLink != "" || newSprRecord.Website != "" ||
		newSprRecord.SpotifyLink != "" || newSprRecord.SoundCloudLink != "" {
		result.Superrare = &newSprRecord
	}

	ch <- result
}

func (f *fetcher) processPoap(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(PoapScanUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processPoap] fetch identity failed"
		ch <- result
		return
	}

	parsedBody, err := fastjson.ParseBytes(body)

	if err != nil {
		result.Err = err
		result.Msg = "[processPoap] parse response json failed"
		ch <- result
		return
	}

	for _, addrToken := range parsedBody.GetArray() {
		eventObj := addrToken.Get("event")

		userPoapIdentity := UserPoapIdentity{
			EventID:    	eventObj.GetInt("id"),
			EventSupply: 	eventObj.GetInt("supply"),
			Year:       	eventObj.GetInt("year"),
			Supply:				addrToken.GetInt("supply"),
			TokenID:    	string(addrToken.GetStringBytes("tokenId")),
			Owner:      	string(addrToken.GetStringBytes("owner")),
			CreatedDate:	string(addrToken.GetStringBytes("created_date")),
			EventDesc:  	string(eventObj.GetStringBytes("description")),
			FancyID:    	string(eventObj.GetStringBytes("fancy_id")),
			EventName:  	string(eventObj.GetStringBytes("name")),
			EventUrl:   	string(eventObj.GetStringBytes("event_url")),
			ImageUrl:   	string(eventObj.GetStringBytes("image_url")),
			Country:    	string(eventObj.GetStringBytes("country")),
			City:       	string(eventObj.GetStringBytes("city")),
			StartDate:  	string(eventObj.GetStringBytes("start_date")),
			EndDate:    	string(eventObj.GetStringBytes("end_date")),
			ExpiryDate: 	string(eventObj.GetStringBytes("expiry_date")),
			DataSource: 	POAP,
		}

		requestQueryParams := map[string]string{
			"query": fmt.Sprintf(`
				{
					event(id: "%d") {
						tokens {
							id
							owner {
								id
							}
						}
					}
				}
			 `, eventObj.GetInt("id")),
		}

		requestQueryBytes, _ := json.Marshal(requestQueryParams)

		body, err := sendRequest(f.httpClient, RequestArgs{
			url:    PoapSubgraphsUrl,
			method: "POST",
			body:   bytes.NewBuffer(requestQueryBytes).Bytes(),
		})

		if err != nil {
			result.Err = err
			result.Msg = "[processPoap] fetch recommendations failed"
			ch <- result
			return
		}

		parsedBody, err := fastjson.ParseBytes(body)

		if err != nil {
			result.Err = err
			result.Msg = "[processPoap] parse recommendation response json failed"
			ch <- result
			return
		}

		var poapRecommendationsResult []PoapRecommendation

		for _, graphToken := range parsedBody.GetArray("data", "event", "tokens") {
			poapRecommendationsResult = append(poapRecommendationsResult, PoapRecommendation{
				TokenID: string(graphToken.GetStringBytes("id")),
				Address: string(graphToken.GetStringBytes("owner", "id")),
				EventID: eventObj.GetInt("id"),
			})
		}

		userPoapIdentity.Recommendations = poapRecommendationsResult
		result.Poap = append(result.Poap, userPoapIdentity)
	}

	ch <- result
	return
}
