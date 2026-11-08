package cases

import "github.com/ultima-pista/ultima-pista/internal/models"

// Catalog holds the launch cases. Solutions never appear in public listings.
var Catalog = []models.Case{
	{
		ID:         "biblioteca-meia-noite",
		Title:      "A biblioteca à meia-noite",
		Theme:      "Ambiente fechado",
		Difficulty: "Fácil",
		PublicText: "A bibliotecária encontra um livro aberto na mesa, uma janela entreaberta e o alarme desativado. Nada foi roubado. Mesmo assim, ela liga para a polícia.",
		Solution: "A bibliotecária mora no andar de cima. À noite, o gato dela entrou pela janela da sala de leitura (deixada entreaberta para ventilação), pulou na mesa e abriu o livro ao se espreguiçar. O alarme estava desativado porque ela mesma o desligou ao descer buscar o animal. Ela chama a polícia porque, ao chegar, viu pegadas molhadas humanas — não do gato — saindo da janela até a rua: alguém havia entrado, visto o gato, e fugiu sem levar nada.",
		Essentials: []string{
			"Havia um gato da bibliotecária no local",
			"O alarme foi desligado por ela ao descer",
			"As pegadas humanas molhadas indicam invasor que fugiu sem roubar",
		},
		Hints: []string{
			"Pergunte quem mais usa aquele espaço à noite.",
			"Nem tudo que está fora do lugar foi tocado por um ladrão.",
			"O alarme desativado pode ser deliberado — e inocente.",
		},
		Illustration: "library",
	},
	{
		ID:         "trem-sem-passageiro",
		Title:      "O trem sem passageiro",
		Theme:      "Viagem",
		Difficulty: "Médio",
		PublicText: "Um revisor encontra um bilhete válido, uma mala aberta e um casaco no assento. O trem não parou desde a estação anterior. Não há passageiro.",
		Solution: "O passageiro era um mágico/ilusionista a caminho de um show. Ele havia combinado com a assistente na estação anterior: ela pegaria a mala de truques pela plataforma enquanto o trem ainda estava parado, e ele sairia pelo outro lado disfarçado com outro casaco (o casaco no assento era parte do número de troca). O bilhete ficou para comprovar a reserva do camarote. O revisor só inspecionou o vagão depois da partida.",
		Essentials: []string{
			"O passageiro saiu na estação anterior, antes da partida",
			"A mala foi retirada de propósito por uma cúmplice",
			"O casaco no assento faz parte de um truque/disfarce",
		},
		Hints: []string{
			"O trem pode não ter parado — mas alguém pode ter saído antes.",
			"Nem toda mala abandonada foi esquecida.",
			"Pense em ofícios que lidam com desaparecimentos no palco.",
		},
		Illustration: "train",
	},
	{
		ID:         "foto-do-casamento",
		Title:      "A foto do casamento",
		Theme:      "Família",
		Difficulty: "Médio",
		PublicText: "Num álbum antigo, uma foto de casamento mostra a noiva sorrindo. O dono do álbum rasga a foto ao meio assim que a vê e guarda só a metade do noivo.",
		Solution: "O dono do álbum é o fotógrafo da cerimônia, não um parente. Na época, ele era apaixonado pela noiva e tirou a foto em segredo para si. Décadas depois, ao reabrir a caixa, reconhece que a noiva na imagem é sua esposa atual — ela nunca lhe contou que já havia casado antes e anulado o primeiro casamento. Ele rasga a metade dela por choque e guarda a do noivo para confrontá-la com a prova sem destruir o rosto do outro homem.",
		Essentials: []string{
			"O dono do álbum é o fotógrafo, não um convidado qualquer",
			"A noiva da foto é a esposa atual dele",
			"Ele não sabia do casamento anterior até ver a foto",
		},
		Hints: []string{
			"Quem costuma aparecer 'atrás' de fotos de casamento?",
			"A reação forte sugere reconhecimento pessoal.",
			"Guardar só o noivo também é uma escolha deliberada.",
		},
		Illustration: "photo",
	},
	{
		ID:         "jardim-congelado",
		Title:      "O jardim congelado",
		Theme:      "Natureza",
		Difficulty: "Difícil",
		PublicText: "Numa manhã de inverno, todas as esculturas do jardim estão cobertas de gelo — exceto uma. O jardineiro é demitido no mesmo dia.",
		Solution: "As 'esculturas' são na verdade figuras de gelo elaboradas para um festival. Durante a noite, uma frente quente derreteu parcialmente o conjunto. O jardineiro, temendo perder o emprego, pulverizou água gelada sobre quase todas para 'salvar' a aparência antes da inspeção matinal. Ele esqueceu uma peça escondida atrás do muro — a única que derreteu de verdade e revelou a fraude. O patrão percebeu o contraste e o demitiu por adulterar o festival.",
		Essentials: []string{
			"As esculturas eram de gelo (festival), não de pedra",
			"O jardineiro regou/gelou quase todas de novo à noite",
			"A peça esquecida derreteu e denunciou a adulteração",
		},
		Hints: []string{
			"Que material muda tanto da noite para o dia no inverno?",
			"Por que alguém cobriria gelo com mais gelo?",
			"A peça intacta pode ser a que ninguém tocou.",
		},
		Illustration: "garden",
	},
	{
		ID:         "carta-sem-remetente",
		Title:      "A carta sem remetente",
		Theme:      "Segredo",
		Difficulty: "Fácil",
		PublicText: "Todo dia 12, uma carta sem remetente chega à mesma casa. Ninguém mora lá há anos. Mesmo assim, a caixa de correio nunca enche.",
		Solution: "O carteiro do bairro é o filho adulto do antigo morador. O pai, antes de morrer, pediu que ele enviasse mensagens anônimas à casa no aniversário da mãe (dia 12), como se ela ainda recebesse notícias do mundo. O próprio carteiro retira as cartas antigas da caixa a cada entrega, para que a ritualidade continue sem atrair a atenção da prefeitura sobre um imóvel abandonado. A caixa nunca enche porque ele a esvazia.",
		Essentials: []string{
			"O carteiro é quem envia e quem esvazia a caixa",
			"O dia 12 marca uma data afetiva familiar",
			"A casa está ligada à família do carteiro",
		},
		Hints: []string{
			"Quem tem acesso garantido à caixa todos os dias?",
			"Uma casa vazia ainda pode significar algo para alguém.",
			"Se a caixa nunca enche, alguém a mantém assim.",
		},
		Illustration: "letter",
	},
	{
		ID:         "relatorio-em-branco",
		Title:      "O relatório em branco",
		Theme:      "Escritório",
		Difficulty: "Médio",
		PublicText: "Na manhã seguinte à auditoria, o diretor encontra sobre a mesa um envelope lacrado com um relatório de vinte páginas — todas em branco. Ele promove a estagiária.",
		Solution: "A auditoria investigava se documentos sensíveis haviam vazado. A estagiária, última a sair, percebeu que a impressora havia falhado e impresso o relatório confidencial só com as marcas de grampeação e a capa — o toner acabara. Em vez de deixar o material ilegível exposto, ela recolheu as páginas, colocou folhas em branco no envelope original (já etiquetado) e trancou o relatório real no cofre da contabilidade, deixando um bilhete interno. O diretor entende: ela protegeu o sigilo e sinalizou o problema sem espalhar o conteúdo. A promoção reconhece o critério.",
		Essentials: []string{
			"O relatório real era confidencial e a impressão falhou",
			"A estagiária protegeu o conteúdo e sinalizou o incidente",
			"O envelope em branco foi um aviso deliberado, não um erro",
		},
		Hints: []string{
			"Páginas em branco podem ser uma mensagem.",
			"Quem ainda estava no escritório após a auditoria?",
			"O diretor parece aliviado — não irritado.",
		},
		Illustration: "office",
	},
}

func PublicList() []models.CasePublic {
	out := make([]models.CasePublic, 0, len(Catalog))
	for _, c := range Catalog {
		out = append(out, models.CasePublic{
			ID:         c.ID,
			Title:      c.Title,
			Theme:      c.Theme,
			Difficulty: c.Difficulty,
		})
	}
	return out
}

func ByID(id string) (*models.Case, bool) {
	for i := range Catalog {
		if Catalog[i].ID == id {
			c := Catalog[i]
			return &c, true
		}
	}
	return nil, false
}

func PublicCase(c *models.Case, includeText bool) models.CasePublic {
	p := models.CasePublic{
		ID:           c.ID,
		Title:        c.Title,
		Theme:        c.Theme,
		Difficulty:   c.Difficulty,
		Illustration: c.Illustration,
	}
	if includeText {
		p.PublicText = c.PublicText
	}
	return p
}
