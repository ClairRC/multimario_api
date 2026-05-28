import requests
from dateutil import parser
import datetime
import warnings
import json
import os
import sys
from dotenv import load_dotenv
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
import models

# Script for getting information from the 602 race sheet and parsing it into a JSON for syncing purposes
# This is filled pretty heavily with magic numbers and assumptions about the sheet, but that's just what happens
# when the main source of truth for the database is a google sheet

load_dotenv()

sheets_api_key = os.getenv("GSHEETS_API_KEY")
spreadsheetId = "1ludkWzuN0ZzMh9Bv1gq9oQxMypttiXkg6AEFvxy_gZk"
dataRange = "A6:E"

ip = "http://localhost"
port = ":8080"

json_out = models.ImportableJSON()

class Record602:
    def __init__(self, playerName=None, sm64Estimate=None, smg1Estimate=None, smsEstimate=None, smg2Estimate=None):
        self.player = playerName
        self.sm64Estimate = sm64Estimate
        self.smg1Estimate = smg1Estimate
        self.smsEstimate = smsEstimate
        self.smg2Estimate = smg2Estimate

    def setName(self, name):
        self.player = name

    def setSM64Estimate(self, estimate):
        self.sm64Estimate = estimate

    def setSMG1Estimate(self, estimate):
        self.smg1Estimate = estimate

    def setSMSEstimate(self, estimate):
        self.smsEstimate = estimate

    def setSMG2Estimate(self, estimate):
        self.smg2Estimate = estimate

def main():
    # Output path
    output_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "syncjson/602update.json")

    # Get sheets values
    r = requests.get(f"https://sheets.googleapis.com/v4/spreadsheets/{spreadsheetId}/values/{dataRange}?key={sheets_api_key}")
    
    # Get 602 date to make sure this is a valid upcoming race
    date = r.json()["range"].split("'")[1]
    now = datetime.date.today().strftime("%Y-%m-%d")
    try:
        date = parser.parse(date, fuzzy=True).strftime("%Y-%m-%d")
    except parser.ParserError:
        print(f"Error: {date} cannot be parsed as a valid date")
        return
    if (date < now):
        raise Exception(f"Error: {date} has already passed")
    
    sheetRecords = getSheetRecords(r.json())
    existingRecords = getExistingRecords(date)
    parseSignups(sheetRecords, existingRecords, date)

    json_out.Export(output_path)

# Takes sheets response json and returns a dict of records in the format {twitchName: record}
def getSheetRecords(responseJSON):
    sheetRecords = {}
    for item in responseJSON["values"]:
        if len(item) == 0:
            break
        newRec = Record602(item[0].lower(), item[1], item[2], item[3], item[4])
        sheetRecords[newRec.player] = newRec

    return sheetRecords

def parseSignups(sheetRecords, existingRecords, date):
    # For each sheet record, if it doesn't exist in existing records, add it
    for name, record in sheetRecords.items():
        if name not in existingRecords:
            # Check if they exist in the database. Add them if not.
            playerRes = requests.get(ip+port+"/players?twitch_name="+name)
            playerRes = playerRes.json()
            if playerRes["success"] == False:
                warnStr = playerRes["error"] + ": " + name
                warnings.warn(warnStr)
                continue
            if len(playerRes["players"]) == 0:
                json_out.NewPlayer(name)
            
            # Add new record for this player
            newRuns = [
                models.Run("sm64_120", record.sm64Estimate),
                models.Run("smg1_120", record.smg1Estimate),
                models.Run("sms_120", record.smsEstimate),
                models.Run("smg2_242", record.smg2Estimate)
            ]
            json_out.NewRecord(name, date, "602", newRuns)
        # If player IS already on the sheet, check for updates
        else:
            updateRuns = []
            if record.sm64Estimate != existingRecords[name].sm64Estimate:
                updateRuns.append(models.Run("sm64_120", record.sm64Estimate))
            if record.smg1Estimate != existingRecords[name].smg1Estimate:
                updateRuns.append(models.Run("smg1_120", record.smg1Estimate))
            if record.smsEstimate != existingRecords[name].smsEstimate:
                updateRuns.append(models.Run("sms_120", record.smsEstimate))
            if record.smg2Estimate != existingRecords[name].smg2Estimate:
                updateRuns.append(models.Run("smg2_242", record.smg2Estimate))
            
            if len(updateRuns) > 0:
                json_out.UpdateRecord(playerName=name, raceDate=date, raceCategory="602", runs=updateRuns)

    # Check if existing record. If they have one that is NOT on the sheet, delete it
    for name, record in existingRecords.items():
        if name not in sheetRecords:
            json_out.DeleteRecord(name, date, "602")

# Takes the date of the race and gets the records for the race in the formate {twitchName: record}
def getExistingRecords(date):
    # Check if race exists. If not, add it
    url = ip + port + "/races?category=602&on="+date
    raceRes = requests.get(url)
    raceRes = raceRes.json()
    if raceRes["success"] == False:
        raise Exception(raceRes["error"])
    if len(raceRes["races"]) == 0:
        # No races, add it to export JSON. Since there are no people in this race, return empty
        json_out.NewRace("602", date)
        return {}
    # Race exists, get the race ID
    raceID = int(raceRes["races"][0]["id"])

    # Get records for this race
    existingRecords = {}
    url = ip + port + "/records/runs?race_id="+str(raceID)
    while True:
        req_records = requests.get(url)
        req_records = req_records.json()
        if req_records["success"] == False:
            raise Exception(req_records["error"])
    
        for run in req_records["runs"]:
            # Get twitch name for this player
            # I should probably fix this in the backend instead of needing to query each of these, but this script runs like every few minutes so for NOW it's fine
            playerName = run["player_name"]
            url = ip + port + "/players?player_name="+playerName
            r = requests.get(url)
            r = r.json()
            if r["success"] == False:
                raise Exception(r["error"])
            playerName = r["players"][0]["twitch_name"]

            category = run["game_category"]
            estimate = run["estimate"]
            if playerName not in existingRecords:
                existingRecords[playerName] = Record602(playerName)
            if category == "sm64_120":
                existingRecords[playerName].setSM64Estimate(estimate)
            elif category == "smg1_120":
                existingRecords[playerName].setSMG1Estimate(estimate)
            elif category == "sms_120":
                existingRecords[playerName].setSMSEstimate(estimate)
            elif category == "smg2_242":
                existingRecords[playerName].setSMG2Estimate(estimate)

        if req_records["meta"]["next_url"] == None:
            break
        url = ip + port + req_records["meta"]["next_url"]
        
    return existingRecords

if __name__ == "__main__":
    main()