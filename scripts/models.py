import json
import os

# Package for parsing race logic into importable json

# Class for runs so they can be easily added    
class Run:
    def __init__(self, category, estimate=None, time=None):
        self.category = category
        self.estimate = estimate
        self.time = time

class ImportableJSON:
    def __init__(self):
        self.json = {}

    # Adds new player to the output JSON
    def NewPlayer(self, newPlayerTwitch: str):
        if "post-players" not in self.json:
            self.json["post-players"] = []
        self.json["post-players"].append({
            "twitch_name": newPlayerTwitch
        })

    # Adds race to the output json
    def NewRace(self, category: str, date: str, status:str="upcoming", start_time:str|None=None):
        if "post-races" not in self.json:
            self.json["post-races"] = []
        self.json["post-races"].append({
            "category": category,
            "date": date,
            "status": status,
            "start_time": start_time
        })

    # Adds new record output to json
    def NewRecord(self, playerName: str, raceDate: str, raceCategory: str, runs: list[Run], finish_time: str|None=None, numCollected: int|None=None):
        if "post-records" not in self.json:
            self.json["post-records"] = []
        newReq = {
            "player_name": playerName,
            "race_date": raceDate,
            "race_category": raceCategory,
            "finish_time": finish_time,
            "num_collected": numCollected,
        }

        if runs is not None:
            newRuns = []
            for run in runs:
                newRuns.append({
                    "category_name": run.category,
                    "time": run.time,
                    "estimate": run.estimate,
                })
            newReq["runs"] = newRuns
        self.json["post-records"].append(newReq)

    # Adds update-records and update-runs
    def UpdateRecord(self, playerName: str, raceDate: str, raceCategory: str, finishTime: str|None=None, numCollected: int|None=None, runs: list[Run]|None = None):
        updateRecords = finishTime != None and numCollected != None
        if "update-records" not in self.json and updateRecords:
            self.json["update-records"] = []
        if updateRecords:
            self.json["update-records"].append({
                "race_date": raceDate,
                "race_category": raceCategory,
                "player_name": playerName,
                "time": finishTime,
                "num_collected": numCollected,
            })

        if runs != None and len(runs) > 0:
            if "update-runs" not in self.json:
                self.json["update-runs"] = []
            for run in runs:
                self.json["update-runs"].append({
                    "player_name": playerName,
                    "race_date": raceDate,
                    "race_category": raceCategory, 
                    "category": run.category,
                    "time": run.time,
                    "estimate": run.estimate,
                })

    # Adds a delete-record to json
    def DeleteRecord(self, playerName: str, raceDate: str, raceCategory: str):
        if "delete-records" not in self.json:
            self.json["delete-records"] = []
        self.json["delete-records"].append({
            "player_name": playerName,
            "race_date": raceDate,
            "race_category": raceCategory,
        })

    # Gets json as string
    def ToString(self) -> str:
        return json.dumps(self.json, indent=4)
    
    # Exports JSON to file
    def Export(self, filePath: str):
        with open(filePath, "w") as f:
            json.dump(self.json, f, indent=4)