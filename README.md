# doggo-photos

A photo uploading repository for users beloved dog photos. It can store any type of photo but photos of dogs are preferred.

Built using Golang, MySQL, Docker, React and some other fun JS libraries.

## Setup

Requires docker be installed locally.

### Backend

```
cd be
docker-compose up --build
```

To init the database run

```
curl  http://localhost:5000/admin/createdb  
```

This is not safe at all but allows for it to be easily reset. Definitely a very bad idea to run in production.

### Frontend
```
cd fe
npm start
```

## Use

Go to localhost:3000 in browser.

There are test images of dogs in `test-data` that can be uploaded
