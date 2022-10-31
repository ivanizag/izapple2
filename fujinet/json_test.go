package fujinet

import (
	"testing"
)

func testQuerys(t *testing.T, message string, queries [][]string) {
	js := NewFnJson()
	errorCode := js.Parse([]uint8(message))

	if errorCode != NoError {
		t.Fatalf("Parse error %v. It should be %v", errorCode, NoError)
	}

	for _, pair := range queries {
		js.Query([]uint8(pair[0]))
		result := string(js.Result)
		if result != pair[1] {
			t.Errorf("Query for %s, returned %s. It should be %s", pair[0], result, pair[1])
		}
	}
}

func TestQueryMap(t *testing.T) {
	// See https://github.com/FujiNetWIFI/fujinet-apps/blob/master/iss-tracker/apple2/src/satellite.c
	testMessage := `{"timestamp": 1667218311, "message": "success", "iss_position": {"latitude": "21.3276", "longitude": "-39.4989"}}`

	testCases := [][]string{
		{"/iss_position/longitude", "-39.4989"},
		{"/iss_position/latitude", "21.3276"},
		{"/timestamp", "1667218311"},
	}
	testQuerys(t, testMessage, testCases)
}

// See https://github.com/FujiNetWIFI/fujinet-apps/blob/master/json-test/atari/jsontest.c
const testArrayMessage = `
[
  {
    "id": "108361296757279278",
    "created_at": "2022-05-25T07:00:06.000Z",
    "in_reply_to_id": null,
    "in_reply_to_account_id": null,
    "sensitive": false,
    "spoiler_text": "",
    "visibility": "public",
    "language": "en",
    "uri": "https://botsin.space/users/osxthemes/statuses/108361286061805267",
    "url": "https://botsin.space/@osxthemes/108361286061805267",
    "replies_count": 0,
    "reblogs_count": 0,
    "favourites_count": 0,
    "edited_at": null,
    "local_only": null,
    "content": "<p>Floppies! - Robert Davis</p>",
    "reblog": null,
    "account": {
      "id": "23439",
      "username": "osxthemes",
      "acct": "osxthemes@botsin.space",
      "display_name": "Macintosh Themes",
      "locked": false,
      "bot": true,
      "discoverable": true,
      "group": false,
      "created_at": "2018-03-28T00:00:00.000Z",
      "note": "<p>I tweet Mac OSX (pre-10.5) and Kaleidoscope (Classic) themes. Bot by <span class=\"h-card\"><a href=\"https://octodon.social/@Eramdam\" class=\"u-url mention\" rel=\"nofollow noopener noreferrer\" target=\"_blank\">@<span>Eramdam</span></a></span>, inspired by kaleidoscopemac@twitter.com. Also on  Twitter at <a href=\"https://twitter.com/osxthemes\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"\">twitter.com/osxthemes</span><span class=\"invisible\"></span></a></p>",
      "url": "https://botsin.space/@osxthemes",
      "avatar": "https://assets.oldbytes.space/assets.oldbytes.space/accounts/avatars/000/023/439/original/322ac0c621b55624.png",
      "avatar_static": "https://assets.oldbytes.space/assets.oldbytes.space/accounts/avatars/000/023/439/original/322ac0c621b55624.png",
      "header": "https://assets.oldbytes.space/assets.oldbytes.space/cache/accounts/headers/000/023/439/original/ea0e0cd513b5a9f7.png",
      "header_static": "https://assets.oldbytes.space/assets.oldbytes.space/cache/accounts/headers/000/023/439/original/ea0e0cd513b5a9f7.png",
      "followers_count": 157,
      "following_count": 1,
      "statuses_count": 17615,
      "last_status_at": "2022-05-25",
      "emojis": [],
      "fields": []
    },
    "media_attachments": [
      {
        "id": "108361296738754794",
        "type": "image",
        "url": "https://assets.oldbytes.space/assets.oldbytes.space/cache/media_attachments/files/108/361/296/738/754/794/original/5785ab0a51d0db1f.gif",
        "preview_url": "https://assets.oldbytes.space/assets.oldbytes.space/cache/media_attachments/files/108/361/296/738/754/794/small/5785ab0a51d0db1f.png",
        "remote_url": "https://files.botsin.space/media_attachments/files/108/361/285/793/211/606/original/7fe52f343cf0c99a.gif",
        "preview_remote_url": null,
        "text_url": null,
        "meta": {
          "original": {
            "width": 213,
            "height": 181,
            "size": "213x181",
            "aspect": 1.1767955801104972
          },
          "small": {
            "width": 213,
            "height": 181,
            "size": "213x181",
            "aspect": 1.1767955801104972
          }
        },
        "description": "Floppies! - Robert Davis",
        "blurhash": "UbLNcMO@QkAAx{jJX4V@8yX9xYX7D@kXoZkV"
      }
    ],
    "mentions": [],
    "tags": [],
    "emojis": [],
    "card": null,
    "poll": null
  }
]`

func TestQueryArray(t *testing.T) {

	testCases := [][]string{
		{"/0/account/display_name", "Macintosh Themes"},
		{"/0/created_at", "2022-05-25T07:00:06.000Z"},
		{"/0/content", "<p>Floppies! - Robert Davis</p>"},
		{"/0/nonexistent", "NULL"},
		{"/1/account/display_name", "NULL"},
		{"/-1/account/display_name", "NULL"},
		{"/zz/account/display_name", "NULL"},
	}
	testQuerys(t, testArrayMessage, testCases)
}
