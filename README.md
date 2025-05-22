# **Back-End Server for "Obsessed with Watermelon"**
## **1. Running flow on your local repo**
### **1.1. Install the following:**
<p align="left">
  <img alt="Top Langs" height="40px" src="https://img.shields.io/badge/Go-%2300ADD8.svg?&logo=go&logoColor=white" />
  <img alt="Top Langs" height="40px" src="https://img.shields.io/badge/Postgres-%23316192.svg?logo=postgresql&logoColor=white" />
  <img alt="Top Langs" height="40px" src="https://img.shields.io/badge/Redis-%23DD0031.svg?logo=redis&logoColor=white" />
</p>

### **1.2. Set up new schema for this app in your local PostgreSQL**
- **1. Setup user:**`sudo -u postgres createuser --interactive`
- **2. Setup new DB:**`sudo -u postgres createdb {DB name whatever you like}`
- **3. Setup password for DB user:**
`sudo -u postgres psql`
`ALTER USER {your_user_name} WITH PASSWORD '{your_password}';`

### **1.3. Place .env file in your local repository with the content below**
```
DB_USER={Your user name for PostgreSQL}
DB_PASSWORD={Your DB password for the user above}
DB_NAME={DB name you set up above on PostgreSQL for this app}
DB_HOST=localhost
DB_SSLMODE=disable
```
Please place this .env file in your root directory for this app.

### **1.4. Run dbMigration.go to set up schema**
Please run dbMigration/dbMigration.go from terminal.
`go run dbMigration/dbMigration.go`

### **1.5. Now you are ready to go!**
You can start the server from a terminal: `go run main.go`

## **2. Directories**
- main.go
- auth/         (Validation of JWT)
- bribe/
  - actions/    (Handle client's actions)
  - broadcast/  (Broadcast game state to clients)
  - connection/ (Manage game instances and websocket connections)
  - database/  （Manage Session ID with Redis）
- handlers/     (Handlers of screen state and websocket connections)
- middleware/   (Manage JWT)
- models/
- screens/      (Handle HTTP requests )
- utils/        (Initialization of Cron and Logger)
