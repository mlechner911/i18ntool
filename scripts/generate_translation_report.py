#!/usr/bin/env python3
import json
import os
from pathlib import Path

root = Path(__file__).resolve().parents[1]
locales_dir = root / 'examples' / 'locales'
reports_dir = root / 'reports'
reports_dir.mkdir(exist_ok=True)

locale_files = sorted([p for p in locales_dir.glob('*.json') if 'backup' not in p.name])
locales = [p.stem for p in locale_files]

# load and flatten
def flatten(d, prefix=''):
    items = {}
    if isinstance(d, dict):
        for k, v in d.items():
            new_key = f"{prefix}.{k}" if prefix else k
            if isinstance(v, dict):
                items.update(flatten(v, new_key))
            else:
                items[new_key] = v
    return items

all_keys = set()
locale_maps = {}
for f in locale_files:
    try:
        data = json.loads(f.read_text())
    except Exception as e:
        print(f"Failed to read {f}: {e}")
        data = {}
    fm = flatten(data)
    locale_maps[f.stem] = fm
    all_keys.update(fm.keys())

all_keys = sorted(all_keys)
# write CSV: key, <locale1>, <locale2>, ...
import csv
out_file = reports_dir / 'translation_report.csv'
with out_file.open('w', newline='') as csvfile:
    w = csv.writer(csvfile)
    header = ['key'] + locales
    w.writerow(header)
    for k in all_keys:
        row = [k]
        for loc in locales:
            row.append(locale_maps.get(loc, {}).get(k, ''))
        w.writerow(row)

print(f"Wrote {out_file}")
