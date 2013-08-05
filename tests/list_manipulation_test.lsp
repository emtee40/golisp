(describe cons
          (== (cons 'a 'b) '(a . b))
          (== (cons 'a '(b c)) '(a b c)))

(describe reverse
          (== (reverse '(a)) '(a))
          (== (reverse '(a b)) '(b a))
          (== (reverse '(a b c d)) '(d c b a)))

(describe flatten
          (== (flatten '(1 2 3 4)) '(1 2 3 4))
          (== (flatten '(1 (2 3) 4)) '(1 2 3 4))
          (== (flatten '(1 (2 (3 4) 5) 6)) '(1 2 (3 4) 5 6)))

(describe flatten*
          (== (flatten* '(1 2 3 4)) '(1 2 3 4))
          (== (flatten* '(1 (2 3) 4)) '(1 2 3 4))
          (== (flatten* '(1 (2 (3 4) 5) 6)) '(1 2 3 4 5 6))
          (== (flatten* '(1 (2 (3 (7 8) 4) 5) 6)) '(1 2 3 7 8 4 5 6)))

(describe partition
          (== (partition 2 '(1 2 3 4 5 6 7 8)) '((1 2) (3 4) (5 6) (7 8)))
          (== (partition 4 '(1 2 3 4 5 6 7 8)) '((1 2 3 4) (5 6 7 8))))
