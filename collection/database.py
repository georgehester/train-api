import psycopg
from typing import Optional
import os

class Database:
    instance: Optional[psycopg.Connection] = None
    initialised: bool = False

    @classmethod
    def initialise(self, host, port, name, username, password) -> None:
        if self.initialised:
            raise RuntimeError("[ Error ][ Database Already Initialised ]")

        self.instance = psycopg.connect(
            host=host,
            port=port,
            dbname=name,
            user=username,
            password=password,
            row_factory=psycopg.rows.dict_row
        )
        self.initialised = True

        file = open(os.path.join(os.path.dirname(__file__), "database", "initialise.sql"), "r")
        script = file.read()
        file.close()

        self.cursor().execute(script)
        self.connection().commit()

    @classmethod
    def connection(self) -> psycopg.Connection:
        if not self.initialised:
            raise RuntimeError("[ Error ][ Database Not Initialised ]")

        return self.instance

    @classmethod
    def cursor(self) -> psycopg.Cursor:
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

    @classmethod
    def execute(self, query: psycopg.abc.QueryNoTemplate, parameters: psycopg.abc.Params | None = None) -> psycopg.Cursor:
        if not self.initialised:
            raise RuntimeError("[ Error ][ Database Not Initialised ]")

        cursor = self.cursor()
        cursor.execute(query, parameters)
        return cursor


    # @classmethod
    # def execute_with_lock(self, sql: str, parameters, lock: threading.Lock) -> sqlite3.Cursor:
    #     if not self.initialised:
    #         raise RuntimeError("[ Error ][ Database Not Initialised ]")

    #     with lock:
    #         return self.cursor().execute(sql, parameters)