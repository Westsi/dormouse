.text
.globl	main
main:
	endbr64
	pushq	%rbp
	movq	%rsp, %rbp
	movl	$2, -4(%rbp)
	movl	-4(%rbp), %eax
	popq	%rbp
	ret
