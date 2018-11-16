(define-derived-mode gowen-mode clojure-mode "Gowen"
  "Major mode for gowen with support for interacting with a
(inferior-lisp) gowen repl."
  (setq-local inferior-lisp-program gowen-command))

(add-to-list 'auto-mode-alist '("\\.gow\\'" . gowen-mode))

(setq gowen-command "go run github.com/niklasfasching/gowen/cmd/gowen")

(defvar gowen-mode-map
  (let ((map (make-sparse-keymap)))
    (set-keymap-parent map clojure-mode-map)

    (define-key map "\C-c\C-x" 'run-lisp)
    (define-key map "\C-x\C-e" 'lisp-eval-last-sexp)
    (define-key map "\M-\C-x"  'lisp-eval-defun)
    (define-key map "\C-c\C-e" 'lisp-eval-defun)
    (define-key map "\C-c\C-r" 'lisp-eval-region)
    (define-key map "\C-c\C-n" 'lisp-eval-form-and-next)
    (define-key map "\C-c\C-p" 'lisp-eval-paragraph)

    (define-key map "\C-c\C-q" 'gowen-inferior-lisp-stop)
    (define-key map "\C-c\M-c" 'gowen-inferior-lisp-start)
    (define-key map "\C-c\M-r" 'gowen-inferior-lisp-restart)
    (define-key map "\C-c\M-o" 'comint-clear-buffer)
    map))

(defun gowen-inferior-lisp-start ()
  (interactive)
  (inferior-lisp gowen-command))

(defun gowen-inferior-lisp-stop ()
  (interactive)
  (let ((process (inferior-lisp-proc)))
    (delete-process process)
    (kill-buffer "*inferior-lisp*")))

(defun gowen-inferior-lisp-restart ()
  (interactive)
  (gowen-inferior-lisp-stop)
  (gowen-inferior-lisp-start))
