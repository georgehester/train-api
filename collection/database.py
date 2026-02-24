import sqlite3
from typing import Optional
import os

class Database:
    instance: Optional[sqlite3.Connection] = None
    initialised: bool = False

    @classmethod
    def initialise(self, path: str) -> None:
        if self.initialised:
            raise RuntimeError("[ Error ][ Database Already Initialised ]")

        self.instance = sqlite3.connect(path, check_same_thread=False)
        self.instance.row_factory = sqlite3.Row
        self.initialised = True

        full_path = os.path.join(os.path.dirname(__file__), "database", "initialise.sql")

        file = open(full_path, "r")
        script = file.read()
        file.close()

        self.cursor().executescript(script)
        self.connection().commit()

    @classmethod
    def connection(self) -> sqlite3.Connection:
        if not self.initialised:
            raise RuntimeError("[ Error ][ Database Not Initialised ]")

        return self.instance

    @classmethod
    def cursor(self) -> sqlite3.Cursor:
        return self.connection().cursor()

    @classmethod
    def close(self) -> None:
        if not self.initialised:
            raise RuntimeError("[ Error ][ Database Not Initialised ]")

        self.instance.close()
        self.instance = None
        self.initialised = False

    @classmethod
    def commit(self) -> None:
        if not self.initialised:
            raise RuntimeError("[ Error ][ Database Not Initialised ]")

        self.connection().commit()