
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


## Local Development
This repo includes a secondary Dockerfile `Dockerfile.dev` specifically
for developing locally against a [sphinx-stack](https://github.com/stakwork/sphinx-stack) environment. There is a docker-compose file included
with this meme-server repo for making sure to connect to the
docker network created by sphinx-stack since meme-server
needs to be able to communicate with those services (e.g. db, aperture, etc.).

The development Dockerfile runs the meme server using [air](https://github.com/cosmtrek/air)
which watches for changes in the code files and recompiles the code on save. 
Having this in a docker container allows the server to interface with the `sphinx-stack`
services. 

Simply run `docker-compose up` in your terminal to run the `air` enabled server. 

You will have to follow the local dev instructions in the 
[sphinx stack readme](https://github.com/stakwork/sphinx-stack/alts/README.md) to make sure
it's running in a way that is compatible with this setup. 

## Optional Paywall for Uploading Large Files
The [Sphinx Stack](https://github.com/stakwork/sphinx-stack) now has support for LSAT-enabled paywalls.
Meme server leverages this for protected routes for uploading large files. 

If you do nothing, meme server will continue to work as before with an upper size limit of 32MB. 

The easiest way to enable the paywall is using docker compose with sphinx stack. This will probably
be simplified in the future, with better docker images, etc., but for now the setup looks like this:

### In Sphinx Stack
Aperture is the proxy server, developed by lightning Labs, that issues LSATs and is the first 
line of defense for validation. This is run by sphinx-stack via docker-compose. 

There is a yaml file used for configuring Aperture. The relevant fields (that can be left at default) are:

```yaml
constraints:
      "large_upload_max_mb": "32"

authwhitelistpaths:
	# update this to be a regex to catch all
	# paths you DON'T want protected
```

If everything is set, you'll want to run sphinx stack _with_ aperture and etcd and _without_ meme:

```shell
$ docker compose -f docker-compose.yml --profile aperture --project-directory . up exclude-services
```

### In Meme 
There are two environment variables we care about, which can  be seen in the local
`docker-compose.yaml` file: `MAX_FREE_UPLOAD_SIZE_MB` (int) and `RESTRICT_UPLOAD_SIZE` (bool).


#### `RESTRICT_UPLOAD_SIZE`
This is our initial "switch" to turn on the protections. If this isn't turned on 
then the default 32MB is used for all file upload paths (except `/largefile` 
will always require an LSAT if aperture is running as a proxy). 
Set this to `1` or `true` to turn on restrictions.

#### `MAX_FREE_UPLOAD_SIZE_MB`
This tells the server how big free uploads should be. This is configurable, but it defaults to 1MB. So if `RESTRICT_UPLOAD_SIZE` is enabled but this is not, 
then upload restrictions will kick in at >1MB.

#### Usage
To turn on the paywall with sphinx-stack running using the commands above, 
simply run the following from the meme directory:

```shell
$ RESTRICT_UPLOAD_SIZE=true MAX_FREE_UPLOAD_SIZE_MB=2 docker compose up
```

(This will also start up a development environment that rebuilds and restarts
the server if a file is changed.)

Now a request `POST /file` will work as before EXCEPT will not allow
any requests with file attachments larger than 2MB. 

Because aperture is running, all requests to `/largefile` will require
an LSAT.  `POST /largefile` with an LSAT and you will be able to upload
anything up to 32MB (which is what aperture sets in the LSAT caveats).

### LSAT Troubleshooting
A couple useful tools for troubleshooting are [Postman](https://postman.com) for
easily creating and editing http requests, and the [LSAT Playground](https://lsat-playground.vercel.app/#caveats)
where you can inspect and manipulate LSATs. 

One interesting thing to try is experimenting with LSAT delegation. 

Take an LSAT that's returned from aperture and paste it into the caveats
section in the playground. You'll see a constraint `large_upload_max_mb=32`.
Now we can add a new constraint to that LSAT: `large_upload_max_mb=3`. Meme
will enforce a restriction that any newer caveats must be increasingly restrictive
such that you couldn't add a caveat where the max is larger than the 32
of the original LSAT.  
The new LSAT macaroon can then be shared with someone else (and is impossible
to revert) and anyone with this LSAT will be restricted to uploads of no
more than 3MB. 