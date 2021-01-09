import { useForm } from "react-hook-form";
import { Row, Col, Button, Form, Alert } from "react-bootstrap";
import axios from "axios";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";

function onCreateAccount(data) {
  console.log(data);
  axios
    .post(`/api/CreateAccount`, {
      username: data.username,
      password: data.password,
    })
    .then((res) => {
      if (res.data) {
        // window.localStorage.setItem("loginKey", res.data.loginKey);
        // setloginKey(res.data.loginKey);
        // setUser(res.data.name);
      }
      console.log(res);
      console.log(res.data);
      //setPage(3);
    });
}

const schema = yup.object().shape({
  username: yup.string().required("Required"),
  password: yup.string().required("Required"),
  confirmPassword: yup
    .string()
    .oneOf([yup.ref("password"), null], "Passwords do not match!")
    .required("Required"),
});

export function CreateAccount() {
  const { register, handleSubmit, watch, errors } = useForm({
    resolver: yupResolver(schema),
  });
  return (
    <Row className="vertical-center">
      <Col md="3"></Col>
      <Col>
        <Form onSubmit={handleSubmit(onCreateAccount)}>
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
          <Form.Group controlId="formBasicPassword">
            <Form.Label>Confirm Password</Form.Label>
            <Form.Control
              name="confirmPassword"
              type="confirmPassword"
              placeholder="Repeat Password"
              ref={register}
            />
          </Form.Group>
          {errors.confirmPassword && (
            <Alert variant="warning">{errors.confirmPassword?.message}</Alert>
          )}
          <Button variant="primary" type="submit">
            Submit
          </Button>
        </Form>
      </Col>
      <Col md="3"></Col>
    </Row>
  );
}
