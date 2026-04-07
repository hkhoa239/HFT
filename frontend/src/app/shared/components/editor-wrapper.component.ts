import { Component, ElementRef, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges, ViewChild } from '@angular/core';
import { EditorState } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { python } from '@codemirror/lang-python';
import { oneDark } from '@codemirror/theme-one-dark';
import { foldGutter, indentOnInput, syntaxHighlighting, defaultHighlightStyle, bracketMatching } from '@codemirror/language';
import { searchKeymap, highlightSelectionMatches } from '@codemirror/search';
import { autocompletion, completionKeymap, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';
import { lintKeymap } from '@codemirror/lint';

@Component({
  selector: 'app-editor',
  standalone: true,
  templateUrl: './editor-wrapper.component.html',
  styleUrl: './editor-wrapper.component.scss'
})
export class EditorWrapperComponent implements OnInit, OnChanges {
  @ViewChild('editorContainer', { static: true }) editorContainer!: ElementRef;
  @Input() value = '';
  @Output() valueChange = new EventEmitter<string>();
  
  private view?: EditorView;

  ngOnInit() {
    this.createEditor();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['value'] && !changes['value'].firstChange && this.view) {
      if (this.view.state.doc.toString() !== this.value) {
        this.view.dispatch({
          changes: { from: 0, to: this.view.state.doc.length, insert: this.value }
        });
      }
    }
  }

  private createEditor() {
    const state = EditorState.create({
      doc: this.value,
      extensions: [
        lineNumbers(),
        highlightActiveLineGutter(),
        history(),
        foldGutter(),
        indentOnInput(),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        bracketMatching(),
        closeBrackets(),
        autocompletion(),
        highlightActiveLine(),
        highlightSelectionMatches(),
        keymap.of([
          ...closeBracketsKeymap,
          ...defaultKeymap,
          ...searchKeymap,
          ...historyKeymap,
          ...completionKeymap,
          ...lintKeymap
        ]),
        python(),
        oneDark,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            this.valueChange.emit(update.state.doc.toString());
          }
        }),
        EditorView.theme({
          "&": { height: "100%", fontSize: "13px" },
          ".cm-scroller": { overflow: "auto", fontFamily: "'Cascadia Code', 'Fira Code', monospace" }
        })
      ]
    });

    this.view = new EditorView({
      state,
      parent: this.editorContainer.nativeElement
    });
  }
}
