#!/usr/bin/env python3
import os
import time
import random
from selenium import webdriver
from selenium.webdriver.firefox.options import Options
from selenium.webdriver.common.by import By
from stem import Signal
from stem.control import Controller
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC


# --- Настройки ---
BRIDGES_URL = "https://bridges.torproject.org/bridges?transport=obfs4&format=text"

LANGS = ["en-US,en", "ru-RU,ru", "fr-FR,fr", "de-DE,de", "en-IN,en"]

TOR_SOCKS_PORT = 9050
TOR_CONTROL_PORT = 9051
TOR_CONTROL_PASS = ""  


def load_user_agents(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        agents = [line.strip() for line in f if line.strip()]
    return agents

# Загрузить список один раз
USER_AGENTS = load_user_agents('/home/sagilyp/GORGONA/bridge-harvester/user_agents.txt')



def renew_tor_ip():
    with Controller.from_port(port=TOR_CONTROL_PORT) as controller:
        controller.authenticate(password=TOR_CONTROL_PASS)
        controller.signal(Signal.NEWNYM)
    print("[*] Tor: NEWNYM signal sent (IP rotated)")
    time.sleep(15)  # подождать смены цепочки

def setup_profile():
    profile = webdriver.FirefoxProfile()
    # Выбрать случайный User-Agent
    random_user_agent = random.choice(USER_AGENTS)
    profile.set_preference("general.useragent.override", random_user_agent)
    profile.set_preference("network.proxy.type", 1)
    profile.set_preference("network.proxy.socks", "127.0.0.1")
    profile.set_preference("network.proxy.socks_port", TOR_SOCKS_PORT)
    profile.set_preference("network.proxy.socks_remote_dns", True)
    profile.set_preference("intl.accept_languages", random.choice(LANGS))
    profile.set_preference("dom.webdriver.enabled", False)
    profile.set_preference('useAutomationExtension', False)
    return profile

def fetch_bridges(vm_name):
    renew_tor_ip()
    options = Options()
    options.headless = True
    profile = setup_profile()
    options.profile = profile
    from selenium.webdriver.firefox.service import Service
    service = Service('/usr/local/bin/geckodriver')
    driver = webdriver.Firefox(service=service, options=options)
    driver.set_window_size(random.randint(900, 1920), random.randint(700, 1200))
    try:
        driver.get(BRIDGES_URL)
        wait = WebDriverWait(driver, 15)
        element = wait.until(EC.presence_of_element_located((By.ID, "bridgelines")))
        bridge_html = element.get_attribute("innerHTML")
        bridge_html = bridge_html.replace('<br>', '\n').replace('<br/>', '\n').replace('<br />', '\n')
        new_bridges = [ln.strip() for ln in bridge_html.split('\n') if ln.strip()]
        new_bridges = set(new_bridges)  # множество новых мостов
        out_dir = os.path.abspath(os.path.dirname(__file__))
        out_file = os.path.join(out_dir, f"bridges_{vm_name}.txt")
        if os.path.exists(out_file):
            with open(out_file, "r") as f:
                existing_bridges = set(line.strip() for line in f if line.strip())
        else:
            existing_bridges = set()
        all_bridges = existing_bridges.union(new_bridges)
        with open(out_file, "w") as f:
            for bridge in sorted(all_bridges):
                f.write(bridge + "\n")
        print(f"[{vm_name}] saved {len(new_bridges)} new bridges, total {len(all_bridges)} unique bridges to {out_file}")
    finally:
        driver.quit()


if __name__ == "__main__":
    import sys
    vm = sys.argv[1] if len(sys.argv) > 1 else "unknown"
    fetch_bridges(vm)
