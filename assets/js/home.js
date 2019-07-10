if ($('#home-page').length > 0) {

    // Players
    $('[data-sort]').on('click', function (e) {

        loadPlayers($(this).attr('data-sort'));
    });

    loadPlayers('level');

    function loadPlayers(sort) {

        $('#players .fa-spin').removeClass('d-none');
        $('#players table').addClass('d-none');

        $('[data-sort]').removeClass('badge-success');
        $('[data-sort="' + sort + '"]').addClass('badge-success');

        $('#column').html(sort);

        $.ajax({
            url: '/home/' + sort + '/players.json',
            dataType: 'json',
            cache: false,
            success: function (data, textStatus, jqXHR) {

                $('#players .fa-spin').addClass('d-none');
                $('#players table').removeClass('d-none');
                $('#players tbody tr').remove();

                if (isIterable(data)) {

                    const $container = $('#players tbody');

                    $container.json2html(
                        data,
                        {
                            '<>': 'tr', 'data-link': '${link}', 'html': [
                                {
                                    '<>': 'td', 'class': 'font-weight-bold', 'html': '${rank}'
                                },
                                {
                                    '<>': 'td', 'class': 'img', 'html': [
                                        {
                                            '<>': 'div', 'class': 'icon-name', 'html': [
                                                {
                                                    '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}'}],
                                                },
                                                {
                                                    '<>': 'div', 'class': 'name', 'html': '${name}'
                                                }
                                            ]
                                        }
                                    ]
                                },
                                {
                                    '<>': 'td', 'html': function () {
                                        return this.value.toLocaleString();
                                    }
                                },
                            ]
                        },
                        {
                            prepend: false,
                        }
                    );

                    observeLazyImages('#players img[data-lazy]');
                }
            },
        });
    }

    // Prices
    $.ajax({
        url: '/home/prices.json',
        dataType: 'json',
        cache: false,
        success: function (data, textStatus, jqXHR) {
            $('#prices .fa-spin').remove();
            $('#prices table').removeClass('d-none');
            if (isIterable(data)) {
                for (const v of data) {
                    addPriceRow(v, false);
                }
            }
        },
    });

    websocketListener('prices', function (e) {

        const data = $.parseJSON(e.data);
        addPriceRow(data.Data, true);
    });

    function addPriceRow(data, addToTop) {

        const $container = $('#prices tbody');

        $container.json2html(
            data,
            {
                '<>': 'tr', 'data-link': '${link}', 'html': [
                    {
                        '<>': 'td', 'class': 'img', 'html': [
                            {
                                '<>': 'div', 'class': 'icon-name', 'html': [
                                    {
                                        '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': '${avatar}'}]
                                    },
                                    {
                                        '<>': 'div', 'class': 'name', 'html': '${name}'
                                    }
                                ],
                            },
                        ]
                    },
                    {
                        '<>': 'td', 'html': '${before}', 'nowrap': 'nowrap'
                    },
                    {
                        '<>': 'td', 'html': '${after}', 'nowrap': 'nowrap'
                    },
                    {
                        '<>': 'td', 'nowrap': 'nowrap', 'html': [
                            {
                                '<>': 'span', 'data-toggle': 'tooltip', 'data-placement': 'left', 'data-livestamp': '${time}',
                            }
                        ],
                    },
                ]
            },
            {
                prepend: addToTop,
            }
        );

        $container.find('tr').slice(15).remove();

        observeLazyImages('#prices img[data-lazy]');
    }
}
