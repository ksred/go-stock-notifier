# A Stock Notifier using Golang

[![Join the chat at https://gitter.im/ksred/go-stock-notifier](https://badges.gitter.im/ksred/go-stock-notifier.svg)](https://gitter.im/ksred/go-stock-notifier?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

_This is a project I am creating to learn the Go language_

The aim of this project is to:
- Ingest stock data from a JSON API
- Save data into a relational database
- Run various analysis on the stock data
- Mail a user with periodic updates on selected stocks
- Mail a user with periodic updates of analysis on stock data

Planned:
- Extend notification system
- Extend analysis systems
- Web frontend
- User accounts

### Installation

Clone this project. 
Copy `config.json.sample` to `config.json` and replace all relevant values. 
Run all files in the `sql/` directory on the database.

MIT License
