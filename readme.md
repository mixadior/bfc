# BFC - bot forms composer

Library designed to generalize logic of managing user input data for bots.
Main goals:

- platform agnostic (telegram examples included)
- one instance for all users with, but separate state for every user
- nested data, e.g `country -> city -> district`
- input order depending from choice, e.g. `categories -> real estate -> living area`, `categories -> autos -> auto brand`
- 3 types of pre-build controls - `text (user input)`, `choice`, `multi-choice`
- yaml serialization
- `previous` / `next` support
- flexible API to manage errors or write intermediate logic, e.g. manage not finished forms, etc. 

# Example 

http://feodosian.com/public/screencast_2023-07-16_16-09-07.mp4 

# Configuration 

Configuration is a bit complex, this is the price for flexibility. API is not stable, pull requests are welcome.
`BFC` designed to be one instance with pre-confgured and in-memory stored data about controls, what fields should be next (order), data(choices).
Data, that different per user session(data entered, progress, customizations) stored in `state` struct.
Main `BFC` composer object is read-only and shouldn't be changed between sessions. You can use
all `BFC` public methods in different threads. 


# Docs 

TBD, look `examples`

All examples has 1 incoming parameter `BOT_TOKEN` environment var. 

| Example                                       | Description                                                                                                                                                                                                  | 
|-----------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| [telegram-bot-api](examples/telegram-bot-api) | Simple BFC form as telegram bot, build on top of [github.com/go-telegram-bot-api/telegram-bot-api/v5](https://github.com/go-telegram-bot-api/telegram-bot-api/v5)                                            |
| [tgbot](examples/tgbot)                       | Simple BFC form as telegram bot, build on top of [github.com/mr-linch/go-tg](https://github.com/mr-linch/go-tg)                                                                                              | 
| [tgbotolxua](examples/tgbotolxua)             | Complex BFC form as telegram bot, it uses `olx.ua` site real estate filters and locations, parsed and serialized into `yaml`. Build on top of [github.com/mr-linch/go-tg](https://github.com/mr-linch/go-tg) | 

# Made in Ukraine 

Glory to Ukraine! 




