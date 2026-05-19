import os
import subprocess
import requests
from groq import Groq

# ── Configurações ──────────────────────────────────────────────────────────────
GROQ_API_KEY  = os.environ["GROQ_API_KEY"]
GITHUB_TOKEN  = os.environ["GITHUB_TOKEN"]
REPO          = os.environ["REPO"]
PR_NUMBER     = os.environ["PR_NUMBER"]
BASE_SHA      = os.environ["BASE_SHA"]
HEAD_SHA      = os.environ["HEAD_SHA"]

# Limite de caracteres do diff enviado ao modelo (evita estourar contexto)
MAX_DIFF_CHARS = 30_000

SYSTEM_PROMPT = """Você é um engenheiro de software sênior fazendo code review.
Analise o diff fornecido e produza um diagnóstico claro em Markdown com as seções:

## ✅ Pontos Positivos
Liste o que está bem feito no código.

## ⚠️ Problemas Encontrados
Para cada problema informe:
- **Severidade**: 🔴 Crítico | 🟡 Aviso | 🔵 Sugestão
- **Arquivo/Linha**: onde ocorre
- **Descrição**: o que está errado e por quê
- **Sugestão de correção**: como corrigir (com trecho de código se necessário)

## 📋 Resumo
Um parágrafo curto com a avaliação geral do PR.

Seja direto e técnico. Não repita o diff na resposta."""


# ── 1. Coletar o diff ──────────────────────────────────────────────────────────
def get_diff() -> str:
    result = subprocess.run(
        ["git", "diff", f"{BASE_SHA}...{HEAD_SHA}"],
        capture_output=True,
        text=True,
        check=True,
    )
    diff = result.stdout
    if len(diff) > MAX_DIFF_CHARS:
        diff = diff[:MAX_DIFF_CHARS] + "\n\n[... diff truncado por ser muito extenso ...]"
    return diff


# ── 2. Chamar o Groq ───────────────────────────────────────────────────────────
def run_review(diff: str) -> str:
    client = Groq(api_key=GROQ_API_KEY)

    response = client.chat.completions.create(
        model="llama-3.3-70b-versatile",
        max_tokens=4096,
        temperature=0.2,
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {
                "role": "user",
                "content": f"Faça o code review do seguinte diff:\n\n```diff\n{diff}\n```",
            },
        ],
    )
    return response.choices[0].message.content


# ── 3. Postar comentário no PR ─────────────────────────────────────────────────
def post_comment(body: str) -> None:
    url = f"https://api.github.com/repos/{REPO}/issues/{PR_NUMBER}/comments"
    headers = {
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    full_body = f"## 🤖 Diagnóstico do Code Review (IA)\n\n{body}"
    resp = requests.post(url, json={"body": full_body}, headers=headers)
    resp.raise_for_status()
    print(f"✅ Comentário postado: {resp.json()['html_url']}")


# ── 4. Deletar comentários antigos do bot (evita spam em re-runs) ──────────────
def delete_old_bot_comments() -> None:
    url = f"https://api.github.com/repos/{REPO}/issues/{PR_NUMBER}/comments"
    headers = {
        "Authorization": f"Bearer {GITHUB_TOKEN}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    resp = requests.get(url, params={"per_page": 100}, headers=headers)
    resp.raise_for_status()
    for comment in resp.json():
        if (
            comment.get("user", {}).get("login") == "github-actions[bot]"
            and "🤖 Diagnóstico do Code Review" in comment.get("body", "")
        ):
            del_resp = requests.delete(
                f"https://api.github.com/repos/{REPO}/issues/comments/{comment['id']}",
                headers=headers,
            )
            del_resp.raise_for_status()
            print(f"🗑️  Comentário antigo removido: {comment['id']}")


# ── Main ───────────────────────────────────────────────────────────────────────
if __name__ == "__main__":
    print("📂 Coletando diff do PR...")
    diff = get_diff()

    if not diff.strip():
        print("⚠️  Diff vazio — nada a revisar.")
        exit(0)

    print(f"📏 Tamanho do diff: {len(diff)} caracteres")

    print("🧹 Removendo comentários antigos...")
    delete_old_bot_comments()

    print("🤖 Executando review com Groq (Llama 3.3)...")
    review = run_review(diff)

    print("💬 Postando diagnóstico no PR...")
    post_comment(review)