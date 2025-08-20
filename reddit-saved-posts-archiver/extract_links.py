#!/usr/bin/env python3
# /// script
# dependencies = ["beautifulsoup4", "lxml"]
# ///

import sys
from bs4 import BeautifulSoup
from dataclasses import dataclass


@dataclass
class Link:
    text: str
    url: str


def extract_links(html_file: str) -> list[Link]:
    ret: list[Link] = []

    with open(html_file, "r", encoding="utf-8") as file:
        soup = BeautifulSoup(file, "lxml")

    links = soup.find_all("a", href=True, attrs={"slot": "full-post-link"})
    for link in links:
        text = link.get_text(strip=True)
        url = link["href"]
        ret.append(
            Link(
                text=text.strip(),
                url="https://reddit.com" + url,
            ),
        )

    return ret


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: uv run extract_links.py <html_file>")
        sys.exit(1)

    html_file: str = sys.argv[1]

    links: list[Link] = extract_links(html_file)

    print("name\tlink")
    for link in links:
        print(f"{link.text}\t{link.url}")
