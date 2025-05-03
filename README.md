# URL Analyzer - Backend

<!-- TOC -->

* [URL Analyzer - Backend](#url-analyzer---backend)
    * [💻 Local Setup Guide](#-local-setup-guide)
        * [Prerequisites](#prerequisites)
        * [Installation](#installation)
    * [🤙 API Documentation](#-api-documentation)
    * [🪐 Deployment](#-deployment)
        * [✅ CI/CD](#-cicd)
    * [🔅 Linked Repositories](#-linked-repositories)
* [Technical Explanations & Validations](#technical-explanations--validations)
    * [🚀 Dependencies Used](#-dependencies-used)
    * [🧪 Testing](#-testing)

<!-- TOC -->

## 💻 Local Setup Guide

### Prerequisites

- Go 1.20 or later (go 1.24.x)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/sendurangr/url-analyzer-api
   cd url-analyzer-api
    ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start the server:
   ```bash
   go run ./cmd/server/main.go
   ```

4. The server will start on `localhost:8080` by default.
   Health Check endpoint is available at `http://localhost:8080/health`

---

## 🤙 API Documentation

- The URL Analyze api is available at `http://localhost:8080/api/v1/url-analyzer?url=<your-url>`.

```bash
curl --request GET \
  --url 'http://localhost:8080/api/v1/url-analyzer?url=https%3A%2F%2Fwww.home24.de%2F'
```

![api-screenshot](./docs/assets/api-screenshot.png)

## 🪐 Deployment

| Services | Endpoints                                                          |
|----------|--------------------------------------------------------------------|
| Frontend | https://d2tiqwdij4sc1n.cloudfront.net                              |
| Backend  | https://8pmmtnd3yw.ap-south-1.awsapprunner.com/api/v1/health-check |

![Infrastructure Diagram](./docs/assets/diagram.svg)

### ✅ CI/CD

- The **CI/CD** pipeline is set up using **GitHub Actions**.
- The pipeline is triggered on every push to the `master` branch.
- The pipeline builds the Docker image and deploys it to AWS App Runner through <br>
  `GitHub Actions -> AWS ECR -> AWS App Runner`.

## 🔅 Linked Repositories

| Services                        | Repositories                                                                                    |
|---------------------------------|-------------------------------------------------------------------------------------------------|
| Backend - Golang                | https://github.com/sendurangr/url-analyzer-api    (Current)                                     |
| Deployment Infrastructure - AWS | https://github.com/sendurangr/url-analyzer-infrastructure <br> (Provisioning through Terraform) |
| Frontend - Angular              | https://github.com/sendurangr/url-analyzer-client                                               |

# Technical Explanations & Validations

## 🚀 Dependencies Used

- `github.com/gin-gonic/gin`
- `github.com/gin-contrib/cors`
- `golang.org/x/net`

## 🧪 Testing [Also in CI/CD]

⚠️ I have not used any external testing libraries like `testify` or `ginkgo` for testing. <br>
I have used the built-in `testing` package for unit and integration tests. <br>
Reasons: built-in testing package is simple and was easy to use for this project. <br>

### Unit Tests

- The project is tested using `go test ./... -cover` command.
- expected output

```bash
PS C:\dev\personal\lucytech\url-analyzer-api> go test ./... -cover
        github.com/sendurangr/url-analyzer-api/cmd/server               coverage: 0.0% of statements
ok      github.com/sendurangr/url-analyzer-api/integrationtest  2.311s  coverage: [no statements]
?       github.com/sendurangr/url-analyzer-api/internal/constants       [no test files]
ok      github.com/sendurangr/url-analyzer-api/internal/handler 0.770s  coverage: 88.2% of statements
        github.com/sendurangr/url-analyzer-api/internal/middleware              coverage: 0.0% of statements
?       github.com/sendurangr/url-analyzer-api/internal/model   [no test files]
        github.com/sendurangr/url-analyzer-api/internal/routes          coverage: 0.0% of statements
ok      github.com/sendurangr/url-analyzer-api/internal/urlanalyzer     1.912s  coverage: 89.8% of statements
        github.com/sendurangr/url-analyzer-api/internal/utils           coverage: 0.0% of statements

```

### Integration Tests

- The integration tests are located in the `integrationtest` package.
- The tests are run using the `go test ./...` command.

## Assumptions

- the webpage's html tags are not malformed or broken.