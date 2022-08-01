"""生成科研项目刷新和掉落数据表
"""
import json
from pathlib import Path

import pandas as pd
import yaml

research4_projects = json.loads(Path('raw/research4_projects.json').read_text())
df = pd.DataFrame(research4_projects['data']['ByProject'])

conf = yaml.safe_load(Path('raw/merge.yml').read_bytes())
duration_map = {str(d).replace('.', ''): d for d in conf['durations']}


def percent_to_float(desc):
    return float(desc[:-1]) / 100


def get_duration(code: str):
    code = code[1:]
    if not code.isnumeric():
        code = code[:-1]
    return duration_map[code]


df['rate'] = df['rate'].apply(percent_to_float)
df['code'] = df['project'].map(conf['projectMap'])
df = df[['project', 'code', 'rate']]


def merge_drop_stats(name, item_map):
    global df
    research4_blueprints = json.loads(Path(f'raw/{name}.json').read_text())
    dfb = pd.DataFrame(research4_blueprints['data']['ByProject'])
    dfb = dfb[dfb['item'].apply(lambda a: a in item_map)]
    dfb['item'] = dfb['item'].map(item_map)

    for name, grp in dfb.groupby('item'):
        grp = grp[['project', 'average']]
        stat_col = 'gain-' + name
        grp.columns = ['project', stat_col + '_raw']
        df = pd.merge(df, grp, how='left', on='project')
        df[stat_col] = df['rate'] * df[stat_col + '_raw']


for sub in conf['merge']:
    merge_drop_stats(**sub)

df = df.groupby(['code']).sum().reset_index()
df['duration'] = df['code'].apply(get_duration)
for col, m in conf['costMap'].items():
    df['cost-' + col] = df['code'].map(m)

# 加权平均值还原为均值
for col in df.columns:
    if col.startswith('gain-'):
        df[col] = df[col] / df['rate']

df = df[conf['resultColumns']].fillna(0)
df.to_csv('data.csv', index=None)
