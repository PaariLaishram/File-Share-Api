package models

//**************** SHARE LINK ****************//
var get_share_link_query = `
	SELECT id, link FROM share_links
`
var add_share_link_query = `
	INSERT INTO share_links(link) VALUES(?)
`

// ****************** USERS ********************
var get_user_query = `SELECT id, email, password FROM users WHERE id IS NOT NULL`
var user_email_filter = ` AND email = ?`

//**************** REFRESH TOKEN  ****************//

var insert_refresh_token_query = ` INSERT INTO refresh_tokens(user, token, jti, expires_at) VALUES(?,?,?,?)`
