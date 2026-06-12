import{Q as x,T as z,aG as j,_ as p,g as q,s as H,a as K,b as Q,t as Z,q as J,l as F,c as X,F as Y,K as ee,a4 as te,e as ae,z as re,H as ne}from"./mermaid.core-BPnNUrQO.js";import{p as ie}from"./chunk-4BX2VUAB-Bo2iW_sx.js";import{p as se}from"./treemap-KMMF4GRG-C372xg7q.js";import{d as L}from"./arc-pD-FI1VU.js";import{o as le}from"./ordinal-Cboi1Yqb.js";import"./index-E9nPRTPp.js";import"./_baseUniq-El2NPsBe.js";import"./_basePickBy-COMTkAx8.js";import"./clone-CRI1sbiQ.js";import"./init-Gi6I4Gst.js";function oe(e,a){return a<e?-1:a>e?1:a>=e?0:NaN}function ce(e){return e}function ue(){var e=ce,a=oe,m=null,y=x(0),s=x(z),o=x(0);function l(t){var n,c=(t=j(t)).length,g,S,v=0,u=new Array(c),i=new Array(c),f=+y.apply(this,arguments),w=Math.min(z,Math.max(-z,s.apply(this,arguments)-f)),h,$=Math.min(Math.abs(w)/c,o.apply(this,arguments)),T=$*(w<0?-1:1),d;for(n=0;n<c;++n)(d=i[u[n]=n]=+e(t[n],n,t))>0&&(v+=d);for(a!=null?u.sort(function(D,C){return a(i[D],i[C])}):m!=null&&u.sort(function(D,C){return m(t[D],t[C])}),n=0,S=v?(w-c*T)/v:0;n<c;++n,f=h)g=u[n],d=i[g],h=f+(d>0?d*S:0)+T,i[g]={data:t[g],index:n,value:d,startAngle:f,endAngle:h,padAngle:$};return i}return l.value=function(t){return arguments.length?(e=typeof t=="function"?t:x(+t),l):e},l.sortValues=function(t){return arguments.length?(a=t,m=null,l):a},l.sort=function(t){return arguments.length?(m=t,a=null,l):m},l.startAngle=function(t){return arguments.length?(y=typeof t=="function"?t:x(+t),l):y},l.endAngle=function(t){return arguments.length?(s=typeof t=="function"?t:x(+t),l):s},l.padAngle=function(t){return arguments.length?(o=typeof t=="function"?t:x(+t),l):o},l}var pe=ne.pie,G={sections:new Map,showData:!1},b=G.sections,N=G.showData,ge=structuredClone(pe),de=p(()=>structuredClone(ge),"getConfig"),fe=p(()=>{b=new Map,N=G.showData,re()},"clear"),me=p(({label:e,value:a})=>{if(a<0)throw new Error(`"${e}" has invalid value: ${a}. Negative values are not allowed in pie charts. All slice values must be >= 0.`);b.has(e)||(b.set(e,a),F.debug(`added new section: ${e}, with value: ${a}`))},"addSection"),he=p(()=>b,"getSections"),ve=p(e=>{N=e},"setShowData"),xe=p(()=>N,"getShowData"),W={getConfig:de,clear:fe,setDiagramTitle:J,getDiagramTitle:Z,setAccTitle:Q,getAccTitle:K,setAccDescription:H,getAccDescription:q,addSection:me,getSections:he,setShowData:ve,getShowData:xe},ye=p((e,a)=>{ie(e,a),a.setShowData(e.showData),e.sections.map(a.addSection)},"populateDb"),Se={parse:p(async e=>{const a=await se("pie",e);F.debug(a),ye(a,W)},"parse")},we=p(e=>`
  .pieCircle{
    stroke: ${e.pieStrokeColor};
    stroke-width : ${e.pieStrokeWidth};
    opacity : ${e.pieOpacity};
  }
  .pieOuterCircle{
    stroke: ${e.pieOuterStrokeColor};
    stroke-width: ${e.pieOuterStrokeWidth};
    fill: none;
  }
  .pieTitleText {
    text-anchor: middle;
    font-size: ${e.pieTitleTextSize};
    fill: ${e.pieTitleTextColor};
    font-family: ${e.fontFamily};
  }
  .slice {
    font-family: ${e.fontFamily};
    fill: ${e.pieSectionTextColor};
    font-size:${e.pieSectionTextSize};
    // fill: white;
  }
  .legend text {
    fill: ${e.pieLegendTextColor};
    font-family: ${e.fontFamily};
    font-size: ${e.pieLegendTextSize};
  }
`,"getStyles"),Ae=we,De=p(e=>{const a=[...e.values()].reduce((s,o)=>s+o,0),m=[...e.entries()].map(([s,o])=>({label:s,value:o})).filter(s=>s.value/a*100>=1).sort((s,o)=>o.value-s.value);return ue().value(s=>s.value)(m)},"createPieArcs"),Ce=p((e,a,m,y)=>{F.debug(`rendering pie chart
`+e);const s=y.db,o=X(),l=Y(s.getConfig(),o.pie),t=40,n=18,c=4,g=450,S=g,v=ee(a),u=v.append("g");u.attr("transform","translate("+S/2+","+g/2+")");const{themeVariables:i}=o;let[f]=te(i.pieOuterStrokeWidth);f!=null||(f=2);const w=l.textPosition,h=Math.min(S,g)/2-t,$=L().innerRadius(0).outerRadius(h),T=L().innerRadius(h*w).outerRadius(h*w);u.append("circle").attr("cx",0).attr("cy",0).attr("r",h+f/2).attr("class","pieOuterCircle");const d=s.getSections(),D=De(d),C=[i.pie1,i.pie2,i.pie3,i.pie4,i.pie5,i.pie6,i.pie7,i.pie8,i.pie9,i.pie10,i.pie11,i.pie12];let E=0;d.forEach(r=>{E+=r});const O=D.filter(r=>(r.data.value/E*100).toFixed(0)!=="0"),M=le(C);u.selectAll("mySlices").data(O).enter().append("path").attr("d",$).attr("fill",r=>M(r.data.label)).attr("class","pieCircle"),u.selectAll("mySlices").data(O).enter().append("text").text(r=>(r.data.value/E*100).toFixed(0)+"%").attr("transform",r=>"translate("+T.centroid(r)+")").style("text-anchor","middle").attr("class","slice"),u.append("text").text(s.getDiagramTitle()).attr("x",0).attr("y",-400/2).attr("class","pieTitleText");const P=[...d.entries()].map(([r,A])=>({label:r,value:A})),k=u.selectAll(".legend").data(P).enter().append("g").attr("class","legend").attr("transform",(r,A)=>{const I=n+c,B=I*P.length/2,V=12*n,U=A*I-B;return"translate("+V+","+U+")"});k.append("rect").attr("width",n).attr("height",n).style("fill",r=>M(r.label)).style("stroke",r=>M(r.label)),k.append("text").attr("x",n+c).attr("y",n-c).text(r=>s.getShowData()?`${r.label} [${r.value}]`:r.label);const _=Math.max(...k.selectAll("text").nodes().map(r=>{var A;return(A=r==null?void 0:r.getBoundingClientRect().width)!=null?A:0})),R=S+t+n+c+_;v.attr("viewBox",`0 0 ${R} ${g}`),ae(v,g,R,l.useMaxWidth)},"draw"),$e={draw:Ce},Pe={parser:Se,db:W,renderer:$e,styles:Ae};export{Pe as diagram};
