package sql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCollection = Collection{
	"Image.SelectFields": Frag(`${alias}.id AS 'image.id', ${alias}.${digest} AS 'image.md5', ${alias}.size AS 'image.size'`),
}

func TestSQL(t *testing.T) {
	// pure
	{
		context := NewContext().
			WithCollection(testCollection)

		stmt, err := Frag(`SELECT * FROM gc_image`).Evaluate(context)

		if assert.NoError(t, err) {
			assert.Equal(t, `SELECT * FROM gc_image`, stmt.GetStmt())
			assert.Empty(t, stmt.GetArgNames())
		}
	}

	// fragment
	{
		ctx := NewContext().
			WithParams(MapParameters{
				"table": "gc_image",
				"id":    1,
				"md5":   "xxx",
			}).
			WithCollection(testCollection)

		e := Frag(`SELECT * FROM ${table} WHERE id = #{id} AND md5 = #{md5}`)

		stmt, err := e.Evaluate(ctx)

		if assert.NoError(t, err) {
			assert.Equal(t, `SELECT * FROM gc_image WHERE id = ? AND md5 = ?`, stmt.GetStmt())
			assert.Len(t, stmt.GetArgNames(), 2)
			assert.Equal(t, "id", stmt.GetArgNames()[0])
			assert.Equal(t, "md5", stmt.GetArgNames()[1])
		}

		ctx.named = true
	}

	// include
	{
		ctx := NewContext().
			WithParams(MapParameters{
				"md5": "md5",
			}).
			WithCollection(testCollection)

		e := Include("Image.SelectFields", false, MapParameters{
			"alias":  "gi",
			"digest": "$md5",
		})

		stmt, err := e.Evaluate(ctx)
		if assert.NoError(t, err) {
			assert.Equal(t, `gi.id AS 'image.id', gi.md5 AS 'image.md5', gi.size AS 'image.size'`, stmt.GetStmt())
		}
	}

	// if
	{
		ctx := NewContext().
			WithParams(MapParameters{
				"test": 1,
			})

		e := If(Test("eq .test 1"), Frag(`SELECT * FROM gc_image`))
		stmt, err := e.Evaluate(ctx)
		if assert.NoError(t, err) {
			assert.Equal(t, `SELECT * FROM gc_image`, stmt.GetStmt())
		}

		e2 := If(Test("eq .test 0"), Frag(`SELECT * FROM gc_image`))
		stmt2, err := e2.Evaluate(ctx)
		if assert.NoError(t, err) {
			assert.Empty(t, stmt2.GetStmt())
		}
	}

	// choose
	{
		e := Choose(
			When(Test(".title"),
				Frag(`AND title like #{title}`)),
			When(Test(".authorName"),
				Frag(`AND author_name like #{authorName}`)),
			OtherWise(
				Frag(`AND featured = 1`)),
		)

		stmt1, err := e.Evaluate(NewContext().
			WithParams(MapParameters{"title": "nihao"}))
		if assert.NoError(t, err) {
			assert.Equal(t, `AND title like ?`, stmt1.GetStmt())
			assert.Len(t, stmt1.GetArgNames(), 1)
			assert.Equal(t, "title", stmt1.GetArgNames()[0])
		}

		stmt2, err := e.Evaluate(NewContext().WithParams(MapParameters{"authorName": "xxx"}))
		if assert.NoError(t, err) {
			assert.Equal(t, `AND author_name like ?`, stmt2.GetStmt())
		}

		stmt3, err := e.Evaluate(NewContext())
		if assert.NoError(t, err) {
			assert.Equal(t, `AND featured = 1`, stmt3.GetStmt())
		}
	}

	// trim
	{
		e := Where(
			If(Test(".state"), Frag(`state = #{state}`)),
			If(Test(".title"), Frag(`AND title like #{title}`)),
			If(Test(".authorName"), Frag(`AND author_name like #{authorName}`)),
		)

		e2 := Set(
			If(Test(".username"), Frag(`username = #{username},`)),
			If(Test(".password"), Frag(`password = #{password},`)),
		)

		{
			stmt, err := e.Evaluate(NewContext().WithParams(MapParameters{"state": 1}))
			if assert.NoError(t, err) {
				assert.Equal(t, `WHERE state = ?`, stmt.GetStmt())
			}
		}

		{
			stmt, err := e.Evaluate(NewContext().WithParams(MapParameters{"title": "'%xxx%'"}))
			if assert.NoError(t, err) {
				assert.Equal(t, `WHERE title like ?`, stmt.GetStmt())
			}
		}

		{
			stmt, err := e.Evaluate(NewContext().WithParams(MapParameters{"authorName": "'%xxx%'"}))
			if assert.NoError(t, err) {
				assert.Equal(t, `WHERE author_name like ?`, stmt.GetStmt())
			}
		}

		{
			stmt, err := e.Evaluate(NewContext().
				WithParams(MapParameters{
					"state":      1,
					"title":      "'%xxx%'",
					"authorName": "'%xxx%'",
				}))
			if assert.NoError(t, err) {
				assert.Equal(t, `WHERE state = ? AND title like ? AND author_name like ?`, stmt.GetStmt())
			}
		}

		{
			stmt, err := e2.Evaluate(NewContext().
				WithParams(MapParameters{
					"password": 1,
					"username": "xxx",
				}))
			if assert.NoError(t, err) {
				assert.Equal(t, `SET username = ?, password = ?`, stmt.GetStmt())
			}
		}
	}

	// Composite
	{
		e := Composite(
			`SELECT c.*,`,
			Trim("", nil, []string{","},
				Include("profileJoinFields", false, MapParameters{"alias": "cp"}),
				Include("bundleJoinFields", false, MapParameters{"alias": "b"}),
			),
			`FROM gc_creation_v2 c
			 	JOIN gc_creation_profile cp ON cp.creation_id = c.id AND cp.stage = #{stage}
    			JOIN gc_creation_bundle b ON cp.bundle_id = b.id AND b.min_platform_version <= #{platformVersion} AND IF(#{platformVersion} > 1, b.min_platform_version > 1, TRUE)`,
			Where(
				If(Test(".idList"),
					`c.id IN (#{idList})`),
				If(Test(".name"),
					`AND c.name LIKE #{name}`),
			),
		)

		stmt, err := e.Evaluate(NewContext().
			WithParams(MapParameters{
				"idList":          []int64{1, 2, 3},
				"name":            "%abc%",
				"stage":           2,
				"platformVersion": 2,
			}).
			WithCollection(Collection{
				"profileJoinFields": Frag(`
					cp.id                  AS 'profile.id',
    				cp.stage               AS 'profile.stage',
    				cp.image_id            AS 'profile.image_id',
    				cp.name                AS 'profile.name',
    				cp.description         AS 'profile.description',
    				cp.bundle_id           AS 'profile.bundle_id',`),
				"bundleJoinFields": Frag(`
					b.id                   AS 'bundle.id',
    				b.path                 AS 'bundle.path',
    				b.version              AS 'bundle.version',
    				b.min_platform_version AS 'bundle.min_platform_version',
    				b.upload_timestamp     AS 'bundle.upload_timestamp',
    				b.publish_timestamp    AS 'bundle.publish_timestamp'`),
			}),
		)

		if assert.NoError(t, err) {
			tokens := strings.Fields(stmt.Stmt)
			res := strings.Join(tokens, " ")
			fmt.Println(res)
		}
	}
}
