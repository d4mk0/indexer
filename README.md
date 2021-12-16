# CyberConnect Indexer

Check out this template to guide you how the identity and connections(followings and followers) info from a variety of open data sources based off an ETH address are fetched.

The basic idea of getting identity is making requests to the corresponding APIs from each platform to collect data and aggregate them into one single struct.
>[Superrare] `https://superrare.com/api/v2/user?address=%s`

Example of the superrare identity structure and its attributes,
```go
type UserSuperrareIdentity struct {
	Username       string
	Homepage       string
	Location       string
	Bio            string
	InstagramLink  string
	TwitterLink    string
	SteemitLink    string
	Website        string
	SpotifyLink    string
	SoundCloudLink string
	DataSource     string
}
```

We could also get some cross-platform data via context api,
>[Context] `https://context.app/api/profile/$address`
```go
case SuperrareContractAddress:
	result.Superrare = &UserSuperrareIdentity{
		Homepage:   entry.Url,
		Username:   entry.Username,
		DataSource: CONTEXT,
	}
```

To retrieve an address's indexed connection list, e.g. on rarible
>[Rarible followings] `https://api-mainnet.rarible.com/marketplace/api/v4/followings?owner=$address`

>[Rarible followers] `https://api-mainnet.rarible.com/marketplace/api/v4/followers?user=$address`


Example of the connection entry structure,
```go
type ConnectionEntryList struct {
	Conn []ConnectionEntry
	Err  error
	msg  string
}

type ConnectionEntry struct {
	From     string
	To       string
	Platform string
}
```

## Interface
```go
type Fetcher interface {
	// fetch following / follower data
	FetchConnections(address string) ([]ConnectionEntry, error)
	// fetch user identity data
	FetchIdentity(address string) (IdentityEntryList, error)
}
```

## Usage

```sh
>> go run main.go
```

## Testing
```sh
>> go test -v test/*/**
```