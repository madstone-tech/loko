# E-Commerce Platform

A classic three-tier web application:

- **Presentation tier**: a React single-page application in the browser.
- **Application tier**: a Go REST API holding all business logic.
- **Data tier**: PostgreSQL with a read replica and daily backups.

Each tier talks only to the one below it.
