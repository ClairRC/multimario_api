import sys
import json
import requests
import warnings
from dotenv import load_dotenv
import os

# Responsible for taking a formatted json and sending POST requests to db
# This is just a script that runs every now and then to add new data so it's not super efficient or clean
# TODO: Finish this. I'm only implementing the methods I need

load_dotenv()

ip = "http://localhost"
port = ":8080"
backend_api_key = os.getenv("MULTIMARIO_API_KEY")

def main():
    if len(sys.argv) < 2:
        print("must provide json file path")
        sys.exit(1)
    
    jsonFilePath = sys.argv[1]
    data = {}

    # Read json
    try: 
        f = open(jsonFilePath, "r")
    except:
        print("Unable to open JSON file")
        sys.exit(1)

    data = json.load(f)

    # Add new players
    if "post-players" in data:
        for newPlayer in data["post-players"]:
            addPlayer(newPlayer["twitch_name"])
    
    # Add new races
    if "post-races" in data: 
        for newRace in data["post-races"]:
            addRace(newRace["category"], newRace["date"], newRace["status"], newRace["start_time"])

    # Add new records
    if "post-records" in data:
        for newRecord in data["post-records"]:
            addRecord(newRecord["race_date"], newRecord["race_category"], newRecord["player_name"], newRecord["finish_time"], newRecord["num_collected"], newRecord["runs"])

    # Delete records
    if "delete-records" in data:
        for record in data["delete-records"]:
            deleteRecord(record["race_date"], record["race_category"], record["player_name"])

    # Update records
    if "update-records" in data:
        for record in data["update-records"]:
            updateRecord(record["race_date"], record["race_category"], record["player_name"], record["time"], record["num_collected"])

    # Update runs
    if "update-runs" in data:
        for run in data["update-runs"]:
            updateRun(run["race_date"], run["race_category"], run["player_name"], run["category"], run["time"], run["estimate"])

# Add player to DB
def addPlayer(twitchName: str):
    url = ip+port+"/players"
    payload = {"twitch_name": twitchName}
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    addPlayerRes = requests.post(url, json=payload, headers=authHead)
    addPlayerRes = addPlayerRes.json()
    if addPlayerRes["success"] == False:
        warnStr = "Unable to add player " + twitchName + ": " + addPlayerRes["error"]
        warnings.warn(warnStr)

# Add race to DB
def addRace(category: str, date: str|None, status: str|None, start_time: str|None):
    url = ip+port+"/races"
    payload = {
        "category": category,
        "date": date,
        "status": status,
        "start_time": start_time,
    }
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    addRaceRes = requests.post(url, json=payload, headers=authHead)
    addRaceRes = addRaceRes.json()
    if addRaceRes["success"] == False:
        warnStr = "Unable to add race: " + addRaceRes["error"]
        warnings.warn(warnStr)

# Add record to DB
def addRecord(raceDate: str, raceCategory: str, playerName: str, finishTime: str|None, numCollected: int|None, runs: list[dict]|None):
    raceID = getRaceID(raceDate, raceCategory)
    url = ip + port + "/records"
    payload = {
        "race_id": raceID,
        "player_name": playerName,
        "finish_time": finishTime,
        "num_collected": numCollected,
        "runs": runs
    }
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    addRecordRes = requests.post(url, json=payload, headers=authHead)
    addRecordRes = addRecordRes.json()
    if addRecordRes["success"] == False:
        warnStr = "Unable to add record: " + addRecordRes["error"]
        warnings.warn(warnStr)

# Updates record in DB
def updateRecord(raceDate: str, raceCategory: str, playerName: str, raceFinishTime: str|None, newNumCollected: int|None):
    raceID = getRaceID(raceDate, raceCategory)
    url = ip + port + "/records?race_id=" + str(raceID) + "&player_name=" + playerName
    payload = {
        "finish_time": raceFinishTime,
        "num_collected": newNumCollected,
    }
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    editRecordRes = requests.patch(url, json=payload, headers=authHead)
    editRecordRes = editRecordRes.json()
    if editRecordRes["success"] == False:
        warnStr = "Unable to edit record: " + editRecordRes["error"]
        warnings.warn(warnStr)

# Updates run in DB
def updateRun(raceDate: str, raceCategory: str, playerName: str, category: str, time: str|None, estimate: str|None):
    raceID = getRaceID(raceDate, raceCategory)
    url = ip + port + "/records/" + str(raceID) + "/" + playerName + "/runs/" + category
    payload = {
        "time": time,
        "estimate": estimate,
    }
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    editRunRes = requests.patch(url, json=payload, headers=authHead)
    editRunRes = editRunRes.json()
    if editRunRes["success"] == False:
        warnStr = "Unable to update run: " + editRunRes["error"]
        warnings.warn(warnStr)

def deleteRecord(raceDate: str, raceCategory: str, playerName: str): 
    raceID = getRaceID(raceDate, raceCategory)
    url = ip + port +"/records/" + str(raceID)+ "/" + playerName
    authStr = "Bearer " + backend_api_key
    authHead = {"Authorization": authStr}

    deleteRecRes = requests.delete(url, headers=authHead)
    deleteRecRes = deleteRecRes.json()
    if deleteRecRes["success"] == False:
        warnStr = "Unable to delete race record: " + deleteRecRes["error"]
        warnings.warn(warnStr)

# Gets race ID from date and category
def getRaceID(date: str, category: str) -> int:
    url = ip + port + "/races?category=" + category + "&on=" + date
    raceRes = requests.get(url)
    raceRes = raceRes.json()
    if raceRes["success"] == False:
        warnStr = "Unable to get race id: " + raceRes["error"]
        warnings.warn(warnStr)
    if len(raceRes["races"]) == 0:
        return -1
    return int(raceRes["races"][0]["id"])

if __name__ == "__main__":
    main()