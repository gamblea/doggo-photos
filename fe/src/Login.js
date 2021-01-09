import { useForm } from "react-hook-form";
import { Row, Col, Button, Form, Alert } from "react-bootstrap";
import axios from "axios";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import { useState } from "react";
import { useHistory } from "react-router-dom";
import Cookies from "js-cookie";

function createOnLogin(setloginKey, setLoginError, history) {
  return (data) => {
    console.log(data);
    axios
      .post(`/api/account/login`, {
        username: data.username,
        password: data.password,
      })
      .then((res) => {
        const loginKey = res.data?.loginKey;
        console.log("loginKey");
        console.log(loginKey);
        if (loginKey) {
          Cookies.set("doggo-photos-loginKey", res.data.loginKey, {
            expires: 7,
          });
          setloginKey(loginKey);
          history.push("/dashboard");
          // window.localStorage.setItem("loginKey", res.data.loginKey);
          // setloginKey(res.data.loginKey);
          // setUser(res.data.name);
        } else {
          const loginError = res.data?.error ? res.data?.error : "Login Error";
          setLoginError(loginError);
        }
        console.log(res);
        console.log(res.data);
      });
  };
}

const schema = yup.object().shape({
  username: yup.string().required("Required"),
  password: yup.string().required("Required"),
});

export function Login({ setloginKey }) {
  const [loginError, setLoginError] = useState("");
  const history = useHistory();
  const { register, handleSubmit, watch, errors } = useForm({
    resolver: yupResolver(schema),
  });
  return (
    <Row className="vertical-center">
      <Col md="3"></Col>
      <Col>
        <Form
          onSubmit={handleSubmit(
            createOnLogin(setloginKey, setLoginError, history)
          )}
        >
          <Form.Group controlId="formBasicEmail">
            <Form.Label>Username</Form.Label>
            <Form.Control
              name="username"
              type="text"
              placeholder="Username"
              ref={register}
            />
          </Form.Group>
          {errors.username && (
            <Alert variant="warning">{errors.username?.message}</Alert>
          )}
          <Form.Group controlId="formBasicPassword">
            <Form.Label>Password</Form.Label>
            <Form.Control
              name="password"
              type="password"
              placeholder="Password"
              ref={register}
            />
          </Form.Group>
          {errors.password && (
            <Alert variant="warning">{errors.password?.message}</Alert>
          )}
          <Button variant="primary" type="submit">
            Submit
          </Button>
        </Form>
        {loginError ? (
          <div style={{ marginTop: "10px" }}>
            <Alert variant="warning">{loginError}</Alert>
          </div>
        ) : null}
      </Col>
      <Col md="3"></Col>
    </Row>
  );
}
