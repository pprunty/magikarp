# EVERYTHING Autonomous Agent (magikarp)

![assets/shiny_magikarp.webp](assets/shiny_magikarp.webp)

## Description

This Python FastAPI application utilizes (just) Dan's data to surface
timely and accurate recommendations, curate personalized entertainment experiences 
and provide a chat companion
using AI and Large Language 
Models (LLMs).

To do this, I scoured the internet for a free version of a reasonably good LLM I could use. The best I found was 
provided by `ollama`. `ollama` allowed me to run a separate server (which runs Meta's Llama 3 LLM) alonside my 
Python FastAPI webserver. Enabling APIs for recommendations, notifications and general AI prompting using Dan's
data.

****
There would be some short look back period for the user's data (i.e 7 days) and this data would be stored on the 
user's operating system / hardware device.

**Human-In-The-Loop:** I also used a general rules based prompt for things which made sense to me, such as:

1. Notify me if my current location is far from my work location and I have a meeting coming up.
2. If I am in a meeting, do not give me personalized recommendations of any kind. 
3. If my current steps average is below my daily average for the time of day, suggest I go for a walk.
4. Do not send more than two notifications less than 30 mins apart unless urgent.

Ideally these rules would be learned by the agent through feedback from global user base and fine-tuned to specific user
based on feedback.

// create api to add to rule engine


The application has two APIs, with documentation in the screenshots below:


## Pre-requisites

- Python 3.12
- Poetry
- ollama (optional / `ollama` image used in docker commands to run the application)
- Docker

_Note: If you wish to download `ollama`, you can follow the instructions [here](https://ollama.com/). However, it is
easier to run the application through Docker by following the instructions in the next section._

## Quickstart

The easiest way to run this application is via docker.

To run the application, make sure you have your Docker daemon running (docker hub desktop app opened) and run the following:

```shell
make docker-build && make docker-run
```

If you wish to contribute to the development of the app, you must
[follow the instructions to download `ollama`](https://ollama.com/) and all other software dependencies outlined in the
pre-requisites section (except for Docker, obviously).

Once you have `ollama` installed on your system, you can run the following:

```shell
make run
```

This command will do the following:

1. Run an `ollama` server, which runs Meta's Llama 3 model using the [Modelfile](./Modelfile) outline.
2. Run a FastAPI web server with some basic APIs for (just) Dan's data.

The FastAPI webserver has two endpoints:

1. **Swagger endpoint:** The Swagger documentation, is available at [https://](http://127.0.0.1:8000/docs). This endpoint 
allows you to test out the APIs.
2. **Redoc endpoint (extra):** The Redoc documentation is a more stylish version of the Swagger documentation which does 
not allow for manual triggering of the APIs.

## What The Application Does

This will run an `ollama` server, which runs Meta's Llama 3 model
