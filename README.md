# Barista (Website)

A website frontend for the Barista blogging platform.

## Usage

1. Install dependencies

- [Go](https://go.dev) (programming language)
- [Node.js v14+](https://nodejs.org/en) (TailwindCSS CLI)
- [make](https://www.gnu.org/software/make/) (build system)

2. Clone the project locally

```sh
git clone https://github.com/LucDeCaf/barista-web-frontend.git
cd barista-web-frontend
```

3. Install project dependencies

```sh
go mod download
```

4. Clone `.env.example` to `.env` and setup environment variables

```env
# .env.example

# Copy this file, rename to .env, and populate with your account details
RECAPTCHA_PROJECT_ID=example-12345678
RECAPTCHA_KEY=0123456789abcdefgABCDEFG
```

5. Build and run the project using `make`

```sh
make
```
