import tkinter as tk
from tkinter import ttk, scrolledtext, messagebox
import threading
import requests
import json
import uuid


class DBWorkbenchClient(tk.Tk):
    def __init__(self):
        super().__init__()
        self.title("DB Workbench Client")
        self.geometry("800x600")

        self.tabControl = ttk.Notebook(self)

        self.tabGet = ttk.Frame(self.tabControl)
        self.tabSearch = ttk.Frame(self.tabControl)
        self.tabQuery = ttk.Frame(self.tabControl)
        self.tabAdd = ttk.Frame(self.tabControl)  # New tab for adding entries

        self.tabControl.add(self.tabGet, text="GET")
        self.tabControl.add(self.tabSearch, text="SEARCH")
        self.tabControl.add(self.tabQuery, text="QUERY")
        self.tabControl.add(self.tabAdd, text="ADD")  # Add the new tab

        self.tabControl.pack(expand=1, fill="both")

        self.setupGetTab()
        self.setupSearchTab()
        self.setupQueryTab()
        self.setupAddTab()  # Setup the new tab for adding entries

    def setupGetTab(self):
        ttk.Label(self.tabGet, text="ID:").grid(column=0, row=0, padx=10, pady=10)
        self.getIDEntry = ttk.Entry(self.tabGet, width=60)
        self.getIDEntry.grid(column=1, row=0, padx=10, pady=10)
        getButton = ttk.Button(self.tabGet, text="Get", command=self.get)
        getButton.grid(column=2, row=0, padx=10, pady=10)
        getAllButton = ttk.Button(self.tabGet, text="Get All", command=self.get_all)
        getAllButton.grid(column=3, row=0, padx=10, pady=10)

        self.getResultText = scrolledtext.ScrolledText(
            self.tabGet, width=70, height=15, wrap=tk.WORD
        )
        self.getResultText.grid(column=0, row=1, columnspan=4, padx=10, pady=10)

    def setupSearchTab(self):
        ttk.Label(self.tabSearch, text="Keyword:").grid(
            column=0, row=0, padx=10, pady=10
        )
        self.searchEntry = ttk.Entry(self.tabSearch, width=60)
        self.searchEntry.grid(column=1, row=0, padx=10, pady=10)
        searchButton = ttk.Button(self.tabSearch, text="Search", command=self.search)
        searchButton.grid(column=2, row=0, padx=10, pady=10)

        self.searchResultText = scrolledtext.ScrolledText(
            self.tabSearch, width=70, height=15, wrap=tk.WORD
        )
        self.searchResultText.grid(column=0, row=1, columnspan=3, padx=10, pady=10)

    def setupQueryTab(self):
        ttk.Label(self.tabQuery, text="Terms (comma-separated):").grid(
            column=0, row=0, padx=10, pady=10
        )
        self.queryEntry = ttk.Entry(self.tabQuery, width=60)
        self.queryEntry.grid(column=1, row=0, padx=10, pady=10)
        queryButton = ttk.Button(self.tabQuery, text="Query", command=self.query)
        queryButton.grid(column=2, row=0, padx=10, pady=10)

        self.queryResultText = scrolledtext.ScrolledText(
            self.tabQuery, width=70, height=15, wrap=tk.WORD
        )
        self.queryResultText.grid(column=0, row=1, columnspan=3, padx=10, pady=10)

    def setupAddTab(self):
        ttk.Label(self.tabAdd, text="Value (JSON):").grid(
            column=0, row=0, padx=10, pady=10
        )
        self.addValueEntry = ttk.Entry(self.tabAdd, width=60)
        self.addValueEntry.grid(column=1, row=0, padx=10, pady=10)

        addButton = ttk.Button(self.tabAdd, text="Add Entry", command=self.add_entry)
        addButton.grid(column=0, row=1, columnspan=2, padx=10, pady=10)

        self.addResultLabel = ttk.Label(self.tabAdd, text="")
        self.addResultLabel.grid(column=0, row=2, columnspan=2, padx=10, pady=10)

    def db_request(self, endpoint, method="get", params=None):
        try:
            if method == "get":
                response = requests.get(endpoint, params=params)
            elif method == "post":  # Add post method for adding entries
                response = requests.post(endpoint, json=params)
            # Add other methods as needed

            if response.status_code == 200:
                return response.json()
            else:
                return {"error": f"Server returned status code {response.status_code}"}
        except Exception as e:
            return {"error": str(e)}

    def get(self):
        id = self.getIDEntry.get()
        endpoint = f"http://localhost:8080/get/{id}"
        threading.Thread(
            target=self.execute_db_request, args=(endpoint, self.getResultText)
        ).start()

    def get_all(self):
        endpoint = "http://localhost:8080/get/all"
        threading.Thread(
            target=self.execute_db_request, args=(endpoint, self.getResultText)
        ).start()

    def search(self):
        keyword = self.searchEntry.get()
        endpoint = f"http://localhost:8080/search/{keyword}"
        threading.Thread(
            target=self.execute_db_request, args=(endpoint, self.searchResultText)
        ).start()

    def query(self):
        terms = self.queryEntry.get().split(",")
        endpoint = f"http://localhost:8080/query/{'/'.join(terms)}"
        threading.Thread(
            target=self.execute_db_request, args=(endpoint, self.queryResultText)
        ).start()

    def add_entry(self):
        value = self.addValueEntry.get()
        endpoint = "http://localhost:8080/add"
        try:
            value_json = json.loads(value)
            if isinstance(value_json, list):
                params = []
                for entry in value_json:
                    entry_id = str(uuid.uuid4())
                    params.append({entry_id: entry})
            else:
                entry_id = str(uuid.uuid4())
                params = [{entry_id: value_json}]
        except json.JSONDecodeError:
            messagebox.showerror("Error", "Invalid JSON input")
            return

        threading.Thread(
            target=self.execute_db_request,
            args=(endpoint, self.addResultLabel, "post", params),
        ).start()

    def execute_db_request(self, endpoint, result_widget, method="get", params=None):
        response = self.db_request(endpoint, method, params)
        if method == "post":
            result_widget.config(text=response.get("message", ""))
        else:
            result_widget.delete("1.0", tk.END)
            self.display_results(response, result_widget)

    def display_results(self, response, result_widget):
        if response is None:
            messagebox.showerror("Error", "No response received from the server.")
            return

        if "error" in response:
            messagebox.showerror("Error", response["error"])
            return

        if isinstance(response, list):
            self.show_list(response, result_widget)
        elif isinstance(response, dict):
            if len(response) == 1:
                key, value = list(response.items())[0]
                self.show_value(key, value, result_widget)
            else:
                self.show_tree(response, result_widget)
                self.show_table(response, result_widget)

    def show_list(self, data, result_widget):
        result_widget.insert(tk.END, "\n".join(data))

    def show_value(self, key, value, result_widget):
        result_widget.insert(tk.END, f"{key}: {value}")

    def show_tree(self, data, result_widget):
        tree = ttk.Treeview(result_widget)
        tree.heading("#0", text="Key")
        self.populate_tree(tree, "", data)
        tree.grid(column=0, row=0, padx=10, pady=10, sticky=(tk.W, tk.E, tk.N, tk.S))

    def populate_tree(self, tree, parent, data):
        if isinstance(data, dict):
            for key, value in data.items():
                if isinstance(value, dict) or isinstance(value, list):
                    node = tree.insert(parent, "end", text=key)
                    self.populate_tree(tree, node, value)
                else:
                    tree.insert(parent, "end", text=key, value=value)
        elif isinstance(data, list):
            for index, item in enumerate(data):
                node = tree.insert(parent, "end", text=str(index))
                self.populate_tree(tree, node, item)
        else:
            tree.insert(parent, "end", text=data)

    def show_table(self, data, result_widget):
        table = ttk.Treeview(result_widget)
        table["columns"] = ["Value"]
        table.heading("#0", text="Key")
        table.heading("Value", text="Value")
        self.populate_table(table, "", data)
        table.grid(column=0, row=0, padx=10, pady=10, sticky=(tk.W, tk.E, tk.N, tk.S))

    def populate_table(self, table, parent, data):
        if isinstance(data, dict):
            for key, value in data.items():
                if isinstance(value, dict) or isinstance(value, list):
                    node = table.insert(parent, "end", text=key, values=[""])
                    self.populate_json_tree(table, node, value)
                else:
                    table.insert(parent, "end", text=key, values=[value])
        elif isinstance(data, list):
            for index, item in enumerate(data):
                node = table.insert(parent, "end", text=str(index), values=[""])
                self.populate_json_tree(table, node, item)
        else:
            table.insert(parent, "end", text="", values=[data])

    def populate_json_tree(self, table, parent, data):
        if isinstance(data, dict):
            for key, value in data.items():
                if isinstance(value, dict) or isinstance(value, list):
                    node = table.insert(parent, "end", text=key, values=[""])
                    self.populate_json_tree(table, node, value)
                else:
                    table.insert(parent, "end", text=key, values=[value])
        elif isinstance(data, list):
            for index, item in enumerate(data):
                node = table.insert(parent, "end", text=str(index), values=[""])
                self.populate_json_tree(table, node, item)
        else:
            table.insert(parent, "end", text="", values=[json.dumps(data)])


app = DBWorkbenchClient()
app.mainloop()
