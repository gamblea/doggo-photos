import { useForm } from "react-hook-form";
import { Row, Col, Button, Form, Alert } from "react-bootstrap";
import axios from "axios";
import { useHistory, Link, useRouteMatch, Route } from "react-router-dom";
import { useEffect, useState } from "react";
import Cookies from "js-cookie";
import Switch from "react-bootstrap/esm/Switch";

export function Dashboard({ loginKey, setloginKey }) {
  const history = useHistory();
  const [upload, setUpload] = useState(false);
  const [username, setUsername] = useState(undefined);
  const [photos, setPhotos] = useState([]);
  useEffect(() => {
    if (loginKey.length > 0) {
      axios
        .post(`/api/account/user`, {
          loginKey: loginKey,
        })
        .then((res) => {
          const username = res.data?.username;
          if (username) {
            setUsername(username);
          } else {
          }
          console.log(res);
          console.log(res.data);
        });
    }
  }, [loginKey]);
  return (
    <div style={{ padding: "5em" }}>
      <Row>
        <Col>
          {username ? (
            <div>
              <Row></Row>
              <Col md="3"></Col>
              <Col>
                <h2 style={{ textAlign: "center" }}>
                  {capitalizeFirstLetter(username)}'s Doggo Photos
                </h2>
              </Col>
              <Col md="3"></Col>

              <Row>
                <Col>
                  <div style={{ textAlign: "center", minHeight: "600px" }}>
                    <div>
                      {upload ? (
                        <UploadPhotos
                          loginKey={loginKey}
                          setUpload={setUpload}
                        />
                      ) : (
                        <DogPhotos
                          photos={photos}
                          setloginKey={setloginKey}
                          setUpload={setUpload}
                          history={history}
                        />
                      )}
                    </div>
                  </div>
                </Col>
                {/* <Col md="6" style={{ textAlign: "center" }}>
                  Photos
                </Col>
                <Col md="6" style={{ textAlign: "center" }}>
                  Adding New Photos
                </Col> */}
              </Row>
            </div>
          ) : (
            <div className="center-block text-center start-menu">
              <p>Please Login</p>
              <Link to="/">
                <Button variant="primary">Login</Button>
              </Link>
            </div>
          )}
        </Col>
      </Row>
    </div>
  );
}
function capitalizeFirstLetter(word) {
  return word.length > 0 ? word.charAt(0).toUpperCase() + word.slice(1) : "";
}

function CreateLogout(setloginKey, history) {
  return () => {
    Cookies.remove("doggo-photos-loginKey");
    setloginKey("");
    history.push("/");
  };
}

// Need to pull photos from db and create links for them
function DogPhotos({ photos, setloginKey, setUpload, history }) {
  return (
    <div>
      <div style={{ minHeight: "400px" }}>
        {photos.length <= 0 ? (
          <p>Add some photos of cute dogs!</p>
        ) : (
          <div></div>
        )}
      </div>
      <Row>
        <Col>
          <Button
            variant="primary"
            style={{ float: "left" }}
            onClick={() => {
              setUpload(true);
            }}
          >
            Upload Photos
          </Button>
        </Col>
        <Col>
          <Button
            variant="primary"
            style={{ float: "right" }}
            onClick={CreateLogout(setloginKey, history)}
          >
            Logout
          </Button>
        </Col>
      </Row>
    </div>
  );
}

function UploadPhotos({ loginKey, setUpload }) {
  return (
    <div>
      <Row>
        <Col md="4"></Col>
        <Col md="4">
          <div
            style={{
              minHeight: "400px",
              margin: "0 auto",
              textAlign: "center",
            }}
          >
            <Form onSubmit={createSubmitUpload(loginKey)}>
              <Form.Group controlId="files">
                <Form.Label>Upload Photos</Form.Label>
                <Form.Control name="photos" type="file" multiple />
              </Form.Group>
              <Button variant="primary" type="submit">
                Submit
              </Button>
            </Form>
          </div>
        </Col>
        <Col md="4"></Col>
      </Row>

      <Button
        style={{ float: "left" }}
        onClick={() => {
          setUpload(false);
        }}
      >
        Back
      </Button>
    </div>
  );
}

function createSubmitUpload(loginKey) {
  return function (event) {
    event.preventDefault();
    const data = new FormData(event.target);

    data.set("loginKey", loginKey);
    axios.post(`/api/photos/upload`, data).then((res) => {
      console.log(res);
    });
  };
}
