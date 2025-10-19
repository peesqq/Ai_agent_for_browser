package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hang-ma/go-browser-agent/internal/browser"
)

type Tools struct{ b *browser.Browser }

func NewTools(b *browser.Browser) *Tools { return &Tools{b: b} }

func (t *Tools) Navigate(ctx context.Context, url string) (string, error) {
	if err := t.b.Goto(url); err != nil {
		return "", err
	}
	t.b.WaitIdle()
	return t.Observe(ctx)
}

func (t *Tools) Click(ctx context.Context, selector string) (string, error) {
	if err := t.b.Page().Click(selector); err != nil {
		return "", err
	}
	t.b.WaitIdle()
	return t.Observe(ctx)
}

func (t *Tools) Type(ctx context.Context, selector, text string, submit bool) (string, error) {
	if err := t.b.Page().Fill(selector, text); err != nil {
		return "", err
	}
	if submit {
		_ = t.b.Page().Keyboard().Press("Enter")
	}
	t.b.WaitIdle()
	return t.Observe(ctx)
}

func (t *Tools) Observe(ctx context.Context) (string, error) {
	js := `
(() => {
  function text(el, max=2000){
    const t = (el.innerText||"").replace(/\s+/g," ").trim();
    return t.slice(0, max);
  }
  const out = [];
  out.push("# URL: " + location.href);
  out.push("# TITLE: " + document.title);
  document.querySelectorAll("h1,h2,h3").forEach(h=>out.push("* "+h.tagName+": "+text(h,300)));
  const inputs=[...document.querySelectorAll("input,textarea,button,select")]
    .slice(0,30).map((e,i)=>({i,tag:e.tagName.toLowerCase(),type:e.type||"",name:e.name||"",placeholder:e.placeholder||"",text:text(e,120)}));
  out.push("# CONTROLS: "+JSON.stringify(inputs));
  const tables=[...document.querySelectorAll("table")].slice(0,2).map(tb=>text(tb,1200));
  if(tables.length) out.push("# TABLES:\\n"+tables.join("\\n---\\n"));
  return out.join("\\n");
})()
`
	content, err := t.b.Page().Evaluate(js)
	if err != nil {
		return "", err
	}
	return content.(string), nil
}

func (t *Tools) SaveArtifacts(prefix string) {
	ts := time.Now().Format("20060102_150405")
	_ = t.b.Screenshot(fmt.Sprintf("run_artifacts/%s_%s.png", prefix, ts))
	_ = t.b.DumpHTML(fmt.Sprintf("run_artifacts/%s_%s.html", prefix, ts))
}
