package search

// PostIndexMapping defines the Elasticsearch mapping for posts
const PostIndexMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "default": {
          "type": "standard"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "long"
      },
      "title": {
        "type": "text",
        "analyzer": "standard",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "content": {
        "type": "text",
        "analyzer": "standard"
      },
      "summary": {
        "type": "text"
      },
      "author_id": {
        "type": "long"
      },
      "author_username": {
        "type": "keyword"
      },
      "circle_id": {
        "type": "long"
      },
      "circle_name": {
        "type": "keyword"
      },
      "status": {
        "type": "keyword"
      },
      "category": {
        "type": "keyword"
      },
      "tags": {
        "type": "keyword"
      },
      "view_count": {
        "type": "integer"
      },
      "hotness_score": {
        "type": "float"
      },
      "published_at": {
        "type": "date"
      },
      "created_at": {
        "type": "date"
      },
      "updated_at": {
        "type": "date"
      }
    }
  }
}
`

// UserIndexMapping defines the Elasticsearch mapping for users
const UserIndexMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "default": {
          "type": "standard"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "long"
      },
      "username": {
        "type": "text",
        "analyzer": "standard",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "email": {
        "type": "keyword"
      },
      "bio": {
        "type": "text"
      },
      "status": {
        "type": "keyword"
      },
      "follower_count": {
        "type": "integer"
      },
      "following_count": {
        "type": "integer"
      },
      "post_count": {
        "type": "integer"
      },
      "created_at": {
        "type": "date"
      }
    }
  }
}
`

const (
	// PostIndex is the name of the posts index
	PostIndex = "posts"
	// UserIndex is the name of the users index
	UserIndex = "users"
)
