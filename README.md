
![Meme Server](https://github.com/stakwork/sphinx-meme/raw/master/sql/meme-server.png)

# sphinx-meme

Decentralized file server for Sphinx memes, attachments, media, and files. Anyone can run a **sphinx-meme** server, and host end-to-end encrypted files for [sphinx.chat](https://sphinx.chat) users.

Sphinx users can automatically auth with **sphinx-meme** by signing a challenge with their LND key. A JWT is returned, allowing access to upload and download files, and search public media.

File access is granted by *mediaTokens* signed by the owner and passed over the Lightning Network. To buy a piece of media, you can send a *purchase* message to the owner, and if you have paid the correct amount, that user will confirm your payment with a *mediaToken* that can only be used by you:

![Purchase Flow](https://github.com/stakwork/sphinx-meme/raw/master/sql/purchase.jpg)

### configuring
```sql
-- use S3 to store files
STORAGE_MODE=s3 
S3_KEY=***
S3_SECRET=***

-- or you can store files locally
STORAGE_MODE=local
LOCAL_ENCRYPTION_KEY=***

-- used in receipt verification
HOST=memes.sphinx.chat

-- key used to sign JWT tokens
JWT_KEY=***

-- Postgres url (AWS RDS env vars can also be used)
DATABASE_URL=***
````

### providing access to a file

Files can be accessed with a *mediaToken* from the file owner. For attachments, the *mediaToken* is simply sent in the attachment message. For purchased media, the *mediaToken* is included in the purchase confirmation message. All media access receipts have an expiration time.

![Media Token](https://github.com/stakwork/sphinx-meme/raw/master/sql/media_token.png)

### authenticating

Authentication is a 3-step process

- GET `/ask` to receive a challenge
```js
// result
{id:'12345',challenge:'67890'}
```

- Sign the challenge with LND "SignMessage" rpc call
```js
// result
{signature:'d3ubh75p45d'} // sig is base64 encoded
```

- POST `/verify` within 10 seconds *(application/x-www-form-urlencoded)*
```js
// body
{id:'12345',sig:'d3ubh75p45d',pubkey:'xxxxx'}
// result
{token:'base64encodedJWT'}
```

The returned token asserts that you are the owner of the pubkey, and lets you upload and manage files. Store token and include in further requests to file server as header: `"Authorization: Bearer {token}"`.

### routes

- GET `/search/{searchTerm}`: Postgres full text search of the file name, description, and tags. Returns an array of files in order of relevancy.

- GET `/file/{mediaToken}`: download the file

- POST `/upload`: the file is hashed with blake2b, and this hash is used as the file key. A SQL record is created with key, name, size, your pubkey, etc.
```js
// Content-Type: multipart/form-data
{
	file: File,
	price: Number, // sats. Can be 0
	ttl: Number, // seconds until receipt expiration after purchase. Default one week
	name: String,
	description: String,
	tags: []String,
	expiry: Number, // optional permanent expiry timestamp
}
```

- POST `/public`: same as above, but file is publically available

- GET `/public/{muid}`: download a public file

- GET `/media/{muid}`: get file info (does not include stats)

**only for file owner:**

- GET `/mymedia`: list all my files

- GET `/mymedia/{muid}`: get file info

**mediaToken**: `{host}.{muid}.{buyerPubKey}.{exp}.{sig}`

- host: domain of meme-server instance
- muid: unique ID of the file content (blake2b hash)
- buyerPubKey: Gives access to the owner of the pubkey
- exp: expiry unix timestamp. (now + ttl)
- sig:secp256k1 signature of double hash of mediaToken. Enables media server to verify that the token was created by the LND key

### notes

- Purchases and receipts are passed as Lightning Network payments, outside of the scope of this server
    - When a purchase is made, merchant node should call **/mymedia/{muid}** to confirm the price/TTL, and check that amount was paid before issuing the *receipt*. Afterward merchant node can call **/purchase/{muid}** to update the stats for that media.
	- Similarly, an attachment message should check TTL before issuing a *mediaToken*
- If a purchase message does not contain the correct amount, the sats should be returned by the merchant node in the *purchase_deny* message
